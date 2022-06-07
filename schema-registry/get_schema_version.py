import requests
import json
def pretty(text):
  print(json.dumps(text, indent=2))

base_uri = "http://localhost:8081"
topic = 'test'

res = requests.get(f'{base_uri}/subjects/{topic}/versions/latest').json()
pretty(res)
