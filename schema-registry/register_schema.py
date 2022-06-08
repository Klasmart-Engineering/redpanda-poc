import requests
import json
def pretty(text):
  print(json.dumps(text, indent=2))

base_uri = "http://localhost:8081"

topic = 'test'

schema = {
  "type": "record",
  "name": f'{topic}',
  "fields": [
    {
      "name": "ID",
      "type": "string",
      "logicalType": "uuid"
    },
    {
      "name": "Value",
      "type": "string"
    }
  ]
}

res = requests.post(
    url=f'{base_uri}/subjects/{topic}/versions',
    data=json.dumps({
      'schema': json.dumps(schema)
    }),
    headers={'Content-Type': 'application/vnd.schemaregistry.v1+json'}).json()
pretty(res)
