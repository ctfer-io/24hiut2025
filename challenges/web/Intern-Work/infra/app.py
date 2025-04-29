from flask import Flask, request, render_template, redirect, session
import os
from db import init_db, get_user

app = Flask(__name__)
app.secret_key = '1dH!TuWyvttD0PBBg8TfTS$%0Kn$!E35QtEydnCA%tapNn2^zm'

@app.route('/', methods=['GET', 'POST'])
def login():
    if request.method == 'POST':
        username = request.form['username']
        password = request.form['password']
        result = get_user(username, password)

        if result:
            session['user'] = result[0]
            return redirect('/dashboard')
        else:
            return render_template('login.html', error='Invalid credentials')

    return render_template('login.html')

@app.route('/robots.txt')
def robots():
    content = (
            "User-agent: *\n"
            "Disallow: /dashboard\n"
            "\n"

            "29/04/2025\n" 
            "----------------------------------\n"
            "# 🚨Pensez à désactiver le compte 'AdminThimothe' à la fin de son stage pour eviter tous problèmes !! 🚨"
    )
    return content, 200, {'Content-Type': 'test/plain'}

@app.route('/dashboard')
def dashboard():
    if 'user' not in session or session['user'] != 'AdminThimothe':
        return redirect('/')
    
    with open('flag.txt', 'r') as f:
        flag = f.read()
    
    return render_template('dashboard.html', flag=flag, total_machines=12, alerts=3, percentage=75)

if __name__ == '__main__':
    init_db()
    app.run(host='0.0.0.0', port=5000)

