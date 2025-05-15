from flask import Flask, request, render_template, redirect, send_file
import json
from os import chdir, path
from random import randint
from string import ascii_letters
from bot import run_thread, stop_thread

chdir(path.dirname(__file__))


app = Flask(__name__, template_folder='templates/', static_folder="node_modules/tinymce", static_url_path="/static")


articles = {}
with open("articles.json") as f:
    articles = json.load(f)


with open("flag.txt") as f:
    FLAG = f.read()


ADMIN_SESSION_KEY = "".join([ascii_letters[randint(0, len(ascii_letters) - 1)] for _ in range(50)])
print(ADMIN_SESSION_KEY)


@app.route("/", methods=["GET"])
def index():
    articles_formated = ""
    for title, content in articles.items():
        articles_formated += f'<div class="article"><h1>{title}</h1><p>{content}</p></div>'
    return render_template("index.html", articles=articles_formated)



@app.route("/create", methods=["GET", "POST"])
def create():
    global articles

    if request.method == "GET":
        return render_template("create.html", message="")

    if request.method == "POST":

        title = request.form.get("title")
        body = request.form.get("tinyMCE")

        if body is None or title is None:
            return render_template("create.html", message='<div class="fail">Body ou title est vide</div>')
        
        if title in articles.keys():
            return render_template("create.html", message='<div class="fail">Un article avec ce titre existe deja</div>')
    
        articles[title] = body

        return render_template("create.html", message='<div class="success">Article cr&eacute;&eacute;</div>')

    return "Bad request"


@app.route("/admin", methods=["GET"])
def admin():

    if request.cookies.get("token") == ADMIN_SESSION_KEY:
        return render_template("admin.html", flag=FLAG)
    
    return redirect("/")



@app.route("/logo.png", methods=["GET"])
def logo():
    return send_file("logo.png")



if __name__ == '__main__':

    print("Staring bot ... ", end="")
    run_thread("http://127.0.0.1:8080/", ADMIN_SESSION_KEY)
    print("done")

    app.run("0.0.0.0", 8080)

    print("Stopping bot ... ", end="")
    stop_thread()
    print("done")
