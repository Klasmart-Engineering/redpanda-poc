import requests
import json
def pretty(text):
  print(json.dumps(text, indent=2))

base_uri = "http://localhost:8081"

res = requests.get(f'{base_uri}/subjects').json()
pretty(res)
