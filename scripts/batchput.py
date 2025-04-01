import boto3
import hashlib

boto3.set_stream_logger(name='botocore', level='DEBUG')
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

#credentials['AccessKeyId']
#credentials['SecretAccessKey']
#credentials['SessionToken']

# Use the temporary credentials to initialize the Kendra client
kendra = boto3.client(
    'kendra',
    region_name='us-east-1',
    aws_access_key_id=credentials['AccessKeyId'],
    aws_secret_access_key=credentials['SecretAccessKey'],
    aws_session_token=credentials['SessionToken']
)
# session = boto3.Session(profile_name="default")
# kendra = session.client("kendra", region_name="us-east-1")


index_id = '636d6c2a-6cd2-4be7-b56b-92d50b5ba2dc'

# Define the S3 path for the PDF
s3_bucket = 'allianceforpeacebuilding-org'
s3_key = '0399.pdf'
s3_url = f's3://{s3_bucket}/{s3_key}'

# Generate a unique document ID based on the S3 path (or optionally download and hash if uniqueness must reflect content)
document_id = hashlib.md5(s3_url.encode()).hexdigest()

print(document_id)

response = kendra.batch_get_document_status(
    IndexId=index_id,
    DocumentInfoList=[
        {'DocumentId': document_id}
    ]
)

print(response)

# # Define the document attributes (metadata)
# attributes = [
#     {'Key': '_file_type', 'Value': {'StringValue': 'PDF'}},
#     {'Key': 'Region', 'Value': {'StringListValue': ['Nepal']}},
#     {'Key': 'Subject_Keywords', 'Value': {'StringListValue': [
#         'safety', 'security', 'security forces', 'community police engagement', 'collaboration', 'research', 'justice'
#     ]}},
#     {'Key': 'source', 'Value': {'StringListValue': ['Alliance for Peacebuilding']}},
#     {'Key': '_authors', 'Value': {'StringListValue': ['Search for Common Ground (SFCG)']}},
#     {'Key': 'Title', 'Value': {'StringValue': 'Community Perceptions of Safety and Security in Dhanusha District'}},
#     {'Key': '_source_uri', 'Value': {'StringValue': s3_url}}
# ]

# # Define the document using S3Path instead of Blob
# document = {
#     'Id': document_id,
#     'S3Path': {
#         'Bucket': s3_bucket,
#         'Key': s3_key
#     },
#     'ContentType': 'PDF',
#     'Attributes': attributes
# }

# #Use the BatchPutDocument API to add the document to the index
# response = kendra.batch_put_document(
#     IndexId=index_id,
#     Documents=[document],
#     RoleArn=role_arn
# )

# # Check for any errors
# if response.get('FailedDocuments'):
#     for failed_doc in response['FailedDocuments']:
#         print(f"Document ID: {failed_doc['Id']}, Error Code: {failed_doc['ErrorCode']}, Error Message: {failed_doc['ErrorMessage']}")
# else:
#     print('Document uploaded successfully.')
