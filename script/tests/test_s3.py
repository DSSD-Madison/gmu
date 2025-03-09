import pytest
from scripts.s3_manager import (
    list_s3_buckets,
    list_s3_files,
)  # Adjust based on project structure
from botocore.exceptions import BotoCoreError, ClientError

TEST_BUCKET = "allianceforpeacebuilding-org"
TEST_BUCKET_1000 = "ipinst-org"  # A real S3 bucket with 1000+ files


def test_list_s3_buckets():
    """Verify that actual S3 buckets are retrievable."""
    try:
        buckets = list_s3_buckets()
        assert isinstance(buckets, list), "Buckets should be a list"
        assert len(buckets) > 0, "No buckets found, check credentials"
        print(f"Buckets retrieved: {buckets}")
    except (BotoCoreError, ClientError) as e:
        pytest.fail(f"AWS S3 request failed: {e}")


def test_list_s3_files():
    """Verify that files are retrievable from an actual S3 bucket."""
    bucket_name = TEST_BUCKET
    files = list_s3_files(bucket_name)

    assert isinstance(files, list), "Returned files should be a list"
    assert len(files) > 0, "No files found, check if the bucket is empty"
    assert all(isinstance(f, str) for f in files), "Each file key should be a string"

    print(f"✅ Retrieved {len(files)} files from {bucket_name}")


def test_list_s3_files_error_handling():
    """Ensure function gracefully handles non-existent or unauthorized buckets."""
    non_existent_bucket = "this-bucket-should-not-exist"
    files = list_s3_files(non_existent_bucket)

    assert files == [], "Function should return an empty list on failure"
    print("✅ Handled missing bucket correctly.")
