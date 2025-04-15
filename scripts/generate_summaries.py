from io import BytesIO
import io
import boto3
from PyPDF2 import PdfReader
import pymupdf 
import os
from PIL import Image
import json
textract_client = boto3.client('textract')

session = boto3.Session(profile_name="default")
s3 = session.client("s3")
pdf_file = s3.get_object(Bucket="bep-json-test-bucket", Key="files/0010.pdf")[
    "Body"
].read()

resource = boto3.resource('s3')

def extract_text_textract(bucket_name, file_key):
    """Extract text from a PDF stored in S3 using Amazon Textract."""
    response = textract_client.start_document_text_detection(
        DocumentLocation={'S3Object': {'Bucket': bucket_name, 'Name': file_key}}
    )

    # Get Job ID
    job_id = response['JobId']

    # Wait for the job to complete
    while True:
        result = textract_client.get_document_text_detection(JobId=job_id)
        status = result['JobStatus']
        if status in ["SUCCEEDED", "FAILED"]:
            break

    if status == "FAILED":
        raise Exception("Textract failed to process document")

    # Extract text
    extracted_text = " ".join(
        [block["Text"] for block in result["Blocks"] if block["BlockType"] == "LINE"]
    )

    print(extracted_text)
    return extracted_text

bedrock_client = boto3.client('bedrock-runtime')

def summarize_with_bedrock(text):
    """Summarize extracted text using Amazon Titan Express."""
    
    prompt = (
        "Summarize the following document in a concise, clear manner:\n\n"
        + text[:4000]  # Trim input to prevent overload
    )

    body = json.dumps({
        "inputText": prompt
    })

    response = bedrock_client.invoke_model(
        modelId="amazon.titan-text-lite-v1",  # Ensure this matches your region
        contentType="application/json",
        body=body
    )

    return json.loads(response["body"].read().decode("utf-8"))["results"][0]["outputText"]


def store_summary_in_s3(bucket_name, summary, file_key):
    """Save the summary as a text file in S3."""
    summary_key = f"summaries/{file_key}.txt"
    s3.put_object(Body=summary, Bucket=bucket_name, Key=summary_key)
    return summary_key

# for bucket in resource.buckets.all():
#     print(bucket.name)
#     for i in bucket.objects.all():
#         print(i.key)
#         if (i.key[-3:] == "pdf"):
#             pdf_text = extract_text_textract(bucket.name, i.key)
#             summary = summarize_with_bedrock(pdf_text)
#             summary_file = store_summary_in_s3(bucket.name, summary, i.key)

test_bucket = resource.Bucket("bep-json-test-bucket")
for i in test_bucket.objects.all():
    print(i.key)
    if (i.key[-3:] == "pdf"):
        pdf_text = extract_text_textract(test_bucket.name, i.key)
        summary = summarize_with_bedrock(pdf_text)
        store_summary_in_s3(test_bucket.name, summary, i.key)
        break



