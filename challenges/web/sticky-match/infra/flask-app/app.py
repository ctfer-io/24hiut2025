import argparse
import re
import threading
import time 
from flask import Flask, request, render_template, jsonify

app = Flask(__name__)
latencies = []
lock = threading.Lock()
UNLOCK_LATENCY_THRESHOLD = 2

@app.route('/')
def home():
    return render_template('home.html')

@app.route('/register', methods=['POST'])
def register():
    user_input = request.form.get('input', '')
    start = time.time()
    match = re.match(r'^(?:[0-9A-Fa-f]{1,8})+(?:-(?:[0-9A-Fa-f]{1,4})+){3}-(?:[0-9A-Fa-f]{1,12})+$', user_input)
    duration = time.time() - start

    with lock:
        latencies.append(duration)

    if not match:
        error_message = f"❌ Please enter a valid GUID."
        return render_template('register.html', user_input=None, message=error_message)

    success_message = f"✅ Registered Bottle: {user_input}"
    return render_template('register.html', user_input=user_input, message=success_message)

@app.route('/stats')
def stats():
    with lock:
        avg = sum(latencies)/len(latencies) if latencies else 0
    return jsonify({"average_latency": avg})

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('-p', '--port', type=str, default=80, help='Port to run the Flask app on')
    args = parser.parse_args()

    app.run(host='0.0.0.0', port=args.port)
