import sqlite3
import os

DB_PATH = 'users.db'

def init_db():
    if not os.path.exists(DB_PATH):
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        c.execute("CREATE TABLE IF NOT EXISTS users (username TEXT, password TEXT)")
        c.execute("DELETE FROM users")
        c.execute("""
            INSERT INTO users (username, password)
            VALUES ('AdminThimothe', 'S*Ba*t&H82r0#s#t3sJZNvE8CWCPe3SgpyfppYQM4EhChpy25d')
        """)
        conn.commit()
        conn.close()

def get_user(username, password):
    conn = sqlite3.connect(DB_PATH)
    c = conn.cursor()
    query = f"SELECT * FROM users WHERE username = '{username}' AND password = '{password}' -- '"
    print(f"[DEBUG] SQL query: {query}")
    try:
        c.execute(query)
        result = c.fetchone()
    except Exception as e:
        print("[!] SQL Error:", e)
        result = None
    conn.close()
    return result
