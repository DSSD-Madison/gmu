import boto3
import hashlib
import urllib.parse

boto3.set_stream_logger(name='botocore', level='ERROR')
role_arn = ''
role_session_name = 'test-session'

# Initialize the STS client to assume the role and get temporary credentials
sts_client = boto3.client('sts', region_name='us-east-1')

# Assume the IAM role
response = sts_client.assume_role(
    RoleArn=role_arn,
    RoleSessionName=role_session_name
)

# Extract the temporary credentials from the assumed role
credentials = response['Credentials']

# Use the temporary credentials to initialize the Kendra client
kendra = boto3.client(
    'kendra',
    region_name='us-east-1',
    aws_access_key_id=credentials['AccessKeyId'],
    aws_secret_access_key=credentials['SecretAccessKey'],
    aws_session_token=credentials['SessionToken']
)

index_id = ''

# Define the S3 path for the PDF
s3_bucket = 'allianceforpeacebuilding-org'
s3_key = '0399.pdf'
s3_url = f's3://{s3_bucket}/{s3_key}'

document_id = s3_url

def convert_s3_uri_to_url(s3_uri: str) -> str:
    if not s3_uri.startswith("s3://"):
        raise ValueError(f"Invalid S3 URI: {s3_uri}")

    uri_parts = s3_uri[len("s3://"):]
    parts = uri_parts.split("/", 1)

    if len(parts) != 2:
        raise ValueError(f"S3 URI format is incorrect: {s3_uri}")

    bucket, file_path = parts
    encoded_path = urllib.parse.quote(file_path)

    return f"https://{bucket}.s3.amazonaws.com/{encoded_path}"

# Define the document attributes (metadata)
attributes = [
    {'Key': '_file_type', 'Value': {'StringValue': 'PDF'}},
    {'Key': 'Region', 'Value': {'StringListValue': ['Nepal']}},
    {'Key': 'Subject_Keywords', 'Value': {'StringListValue': [
        'safety', 'security', 'security forces', 'community police engagement', 'collaboration', 'research', 'justice'
    ]}},
    {'Key': 'source', 'Value': {'StringListValue': ['Alliance for Peacebuilding']}},
    {'Key': '_authors', 'Value': {'StringListValue': ['Search for Common Ground (SFCG)']}},
    {'Key': 'Title', 'Value': {'StringValue': 'Community Perceptions of Safety and Security in Dhanusha District'}},
    {'Key': '_source_uri', 'Value': {'StringValue': convert_s3_uri_to_url(s3_url)}}
]

# Define the document using S3Path instead of Blob
document = {
    'Id': document_id,
    'S3Path': {
        'Bucket': s3_bucket,
        'Key': s3_key
    },
    'ContentType': 'PDF',
    'Attributes': attributes,
    'Title': 'Community Perceptions of Safety and Security in Dhanusha District',
}

#Use the BatchPutDocument API to add the document to the index
response = kendra.batch_put_document(
    IndexId=index_id,
    Documents=[document],
    RoleArn=role_arn
)
print(response)

'''Some code is below for error handling and document deletion.'''

# if response.get('FailedDocuments'):
#     for failed_doc in response['FailedDocuments']:
#         print(f"Document ID: {failed_doc['Id']}, Error Code: {failed_doc['ErrorCode']}, Error Message: {failed_doc['ErrorMessage']}")
# else:
#     print('Document uploaded successfully.')


# response = kendra.batch_get_document_status(
#     IndexId=index_id,
#     DocumentInfoList=[
#         {'DocumentId': document_id}
#     ]
# )
#
# response = kendra.batch_delete_document(
#     IndexId=index_id,
#     DocumentIdList=[document_id]
# )
