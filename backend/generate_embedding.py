import os
import openai
import psycopg2
import json
from pgvector.psycopg2 import register_vector

# Get API key from environment
api_key = os.popen("grep OPENAI_API_KEY .env | cut -d'=' -f2").read().strip()
openai.api_key = api_key

# Connect to database
conn = psycopg2.connect("postgres://engramiq:engramiq_dev_2024@localhost:5432/engramiq")
register_vector(conn)
cur = conn.cursor()

# Get document
cur.execute("SELECT id, processed_content FROM documents WHERE id='e9eb2db2-29d9-480c-bbd1-b1666aa515c0'")
doc_id, content = cur.fetchone()

print(f"Generating embedding for document {doc_id}")
print(f"Content length: {len(content)} characters")

# Generate embedding
response = openai.Embedding.create(
    input=content[:8000],  # Limit content length
    model="text-embedding-ada-002"
)
embedding = response['data'][0]['embedding']
print(f"Generated embedding with {len(embedding)} dimensions")

# Update document
cur.execute(
    "UPDATE documents SET embedding = %s WHERE id = %s",
    (embedding, doc_id)
)
conn.commit()
print("Document embedding updated successfully")

cur.close()
conn.close()
