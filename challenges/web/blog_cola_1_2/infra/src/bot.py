from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.chrome.service import Service
from time import sleep
from threading import Thread

def visit(url, token):
    chrome_options = Options()
    chrome_options.add_argument("--headless")
    chrome_options.add_argument("--incognito")
    chrome_options.add_argument("--no-sandbox")
    chrome_options.add_argument("--disable-gpu")
    chrome_options.add_argument("--disable-jit")
    chrome_options.add_argument("--disable-wasm")
    chrome_options.add_argument("--disable-dev-shm-usage")
    chrome_options.add_argument("--ignore-certificate-errors")
    chrome_options.page_load_strategy = "eager"
    chrome_options.binary_location = "/usr/bin/chromium-browser"

    service = Service("/usr/bin/chromedriver")
    driver = webdriver.Chrome(service=service, options=chrome_options)

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
    except:pass

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
            sleep(30) # wait 30s




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
        print(f"An error already occured ! {e}")