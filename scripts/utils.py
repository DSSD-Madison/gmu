import os
import psycopg2
import boto3
from dotenv import load_dotenv
from urllib.parse import urlparse

# Load environment variables
load_dotenv()

# Database config
DB_HOST = os.getenv('DB_HOST')
DB_NAME = os.getenv('DB_NAME')
DB_USER = os.getenv('DB_USER')
DB_PASSWORD = os.getenv('DB_PASSWORD')

# Role assumption config
ROLE_ARN = os.getenv('ROLE_ARN')

def get_db_connection():
    return psycopg2.connect(
        host=DB_HOST,
        dbname=DB_NAME,
        user=DB_USER,
        password=DB_PASSWORD
    )

def parse_s3_uri(uri):
    parsed = urlparse(uri)
    return parsed.netloc, parsed.path.lstrip("/")

def assume_role(role_session_name):
    sts = boto3.client('sts', region_name='us-east-1')
    creds = sts.assume_role(
        RoleArn=ROLE_ARN,
        RoleSessionName=role_session_name
    )['Credentials']
    return creds

def get_s3_client(role_session_name):
    creds = assume_role(role_session_name)
    return boto3.client(
        's3',
        region_name='us-east-1',
        aws_access_key_id=creds['AccessKeyId'],
        aws_secret_access_key=creds['SecretAccessKey'],
        aws_session_token=creds['SessionToken']
    )

def get_s3_resource(role_session_name):
    creds = assume_role(role_session_name)
    return boto3.resource(
        's3',
        region_name='us-east-1',
        aws_access_key_id=creds['AccessKeyId'],
        aws_secret_access_key=creds['SecretAccessKey'],
        aws_session_token=creds['SessionToken']
    )

def get_kendra_client(role_session_name):
    creds = assume_role(role_session_name)
    return boto3.client(
        'kendra',
        region_name='us-east-1',
        aws_access_key_id=creds['AccessKeyId'],
        aws_secret_access_key=creds['SecretAccessKey'],
        aws_session_token=creds['SessionToken']
    )
