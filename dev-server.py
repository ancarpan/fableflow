#!/usr/bin/env python3
"""
Development server for FableFlow
Serves static files and proxies API requests to the Go backend
"""

import http.server
import socketserver
import urllib.request
import urllib.parse
import json
import os
import signal
import socket
import sys
from pathlib import Path

class FableFlowDevHandler(http.server.SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, directory=str(Path(__file__).parent / "frontend"), **kwargs)
    
    def do_GET(self):
        # Handle API requests - proxy to backend
        if self.path.startswith('/api/'):
            self.proxy_to_backend()
        # Handle reader routes - proxy to backend
        elif self.path.startswith('/read/'):
            self.proxy_to_backend()
        # Handle static files
        elif self.path.startswith('/static/'):
            super().do_GET()
        # Handle root and other routes - serve index.html
        else:
            self.serve_index()
    
    def do_POST(self):
        # Handle API requests - proxy to backend
        if self.path.startswith('/api/'):
            self.proxy_to_backend()
        else:
            self.send_error(404)
    
    def do_OPTIONS(self):
        # Handle CORS preflight requests
        self.send_response(200)
        self.send_cors_headers()
        self.end_headers()
    
    def proxy_to_backend(self):
        """Proxy requests to the Go backend"""
        try:
            # Construct backend URL
            backend_url = f"http://localhost:8080{self.path}"
            
            # Prepare request
            req_data = None
            if hasattr(self, 'rfile') and self.rfile:
                content_length = int(self.headers.get('Content-Length', 0))
                if content_length > 0:
                    req_data = self.rfile.read(content_length)
            
            # Create request
            req = urllib.request.Request(
                backend_url,
                data=req_data,
                method=self.command,
                headers=dict(self.headers)
            )
            
            # Make request to backend
            with urllib.request.urlopen(req) as response:
                # Send response headers
                self.send_response(response.getcode())
                self.send_cors_headers()
                
                # Copy headers from backend response
                for header, value in response.headers.items():
                    if header.lower() not in ['content-encoding', 'content-length', 'transfer-encoding']:
                        self.send_header(header, value)
                self.end_headers()
                
                # Copy response body
                self.wfile.write(response.read())
                
        except Exception as e:
            print(f"Error proxying to backend: {e}")
            self.send_error(502, f"Backend connection failed: {e}")
    
    def serve_index(self):
        """Serve the main index.html file"""
        try:
            index_path = Path(__file__).parent / "frontend" / "templates" / "index.html"
            if index_path.exists():
                with open(index_path, 'rb') as f:
                    content = f.read()
                self.send_response(200)
                self.send_header('Content-Type', 'text/html')
                self.send_cors_headers()
                self.end_headers()
                self.wfile.write(content)
            else:
                self.send_error(404, "index.html not found")
        except Exception as e:
            self.send_error(500, f"Error serving index: {e}")
    
    def send_cors_headers(self):
        """Send CORS headers"""
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS')
        self.send_header('Access-Control-Allow-Headers', 'Content-Type, Authorization')

def main():
    PORT = 3000
    print(f"🚀 Starting FableFlow development server on port {PORT}")
    print(f"📁 Serving static files from: {Path(__file__).parent / 'frontend'}")
    print(f"🔄 Proxying API requests to: http://localhost:8080")
    print(f"🌐 Frontend: http://localhost:{PORT}")
    print(f"🔧 Backend: http://localhost:8080")
    print("")
    print("Press Ctrl+C to stop")
    
    # Create server with proper socket options
    class ReusableTCPServer(socketserver.TCPServer):
        allow_reuse_address = True
        
        def server_bind(self):
            self.socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            super().server_bind()
    
    # Set up signal handling
    def signal_handler(sig, frame):
        print("\n🛑 Development server stopped")
        sys.exit(0)
    
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    
    try:
        with ReusableTCPServer(("", PORT), FableFlowDevHandler) as httpd:
            httpd.serve_forever()
    except KeyboardInterrupt:
        print("\n🛑 Development server stopped")
    except OSError as e:
        if e.errno == 48:  # Address already in use
            print(f"❌ Port {PORT} is already in use. Trying to kill existing processes...")
            os.system(f"lsof -ti:{PORT} | xargs kill -9 2>/dev/null || true")
            print("🔄 Retrying in 2 seconds...")
            import time
            time.sleep(2)
            # Retry once
            try:
                with ReusableTCPServer(("", PORT), FableFlowDevHandler) as httpd:
                    httpd.serve_forever()
            except OSError:
                print(f"❌ Still unable to bind to port {PORT}. Please check for other processes.")
                sys.exit(1)
        else:
            raise

if __name__ == "__main__":
    main()
