import argparse
import os
import signal
import subprocess
import time
from flask import Flask, render_template

TIMEOUT = 10
FLAG = "24HIUT{^bu1ld_r0bust_r3gex$}"

flag_app = Flask(__name__)

@flag_app.route('/')
@flag_app.route('/register')
def flag():
    return render_template('flag.html', flag=FLAG)

def monitor_website(exposed_port: int):
    # Start the monitored Flask app
    monitored_url = f'http://127.0.0.1:{exposed_port}'
    flask_process = subprocess.Popen(["python3", "app.py", "--port", exposed_port], preexec_fn=os.setsid)
    time.sleep(2)

    try:
        while True:
            try:
                # Use curl as a healthcheck
                response = subprocess.run(
                    ["curl", "-m", str(TIMEOUT), monitored_url],
                    stdout=subprocess.PIPE,
                    stderr=subprocess.PIPE,
                )
                if response.returncode != 0:
                    print("Website is unresponsive. Terminating monitored app...")
                    break
                print("Website is responsive. Monitoring continues...")
            except Exception as e:
                print(f"Error during monitoring: {e}")
                break

            # Healthcheck every 5 seconds
            time.sleep(5)

    finally:
        # Kill the monitored Flask app
        os.killpg(os.getpgid(flask_process.pid), signal.SIGTERM)
        print("Monitored app terminated.")

        # Start the flag-exposing Flask app
        print("Starting flag-exposing app...")
        flag_app.run(host="0.0.0.0", port=exposed_port)

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('-p', '--port', type=str, default=80, help='Port to run the Flask app on')
    args = parser.parse_args()
    monitor_website(args.port)