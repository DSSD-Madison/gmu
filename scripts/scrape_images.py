from io import BytesIO
import io
import boto3
from PyPDF2 import PdfReader
import pymupdf 
import os
from PIL import Image

session = boto3.Session(profile_name="default")
s3 = session.client("s3")
pdf_file = s3.get_object(Bucket="bep-json-test-bucket", Key="files/0010.pdf")[
    "Body"
].read()

resource = boto3.resource('s3')
for bucket in resource.buckets.all():
    print(bucket.name)

# while True:
#     try:
#         pdf_file = s3.get_object(Bucket="bep-json-test-bucket", Key="files/0010.pdf")[
#             "Body"
#         ].read()
#     except:
#         break

test_bucket = resource.Bucket("bep-json-test-bucket")
for i in test_bucket.objects.all():
    if (i.key[-3:] == "pdf"):
            file = s3.get_object(Bucket=test_bucket.name, Key=i.key)["Body"].read()
            pdf_stream = io.BytesIO(file)
            pdf_document = pymupdf.open(stream=pdf_stream, filetype="pdf")

            page = pdf_document[0]
            pix = page.get_pixmap()
            img_data = pix.tobytes("png")

            image = Image.open(io.BytesIO(img_data))
            print(image.size)
            size = (10000, 120)
            image.thumbnail(size)
            #image.show()
            print(type(image))

            bytes = BytesIO()
            image.save(bytes, format="webp")
            webp = i.key[:-3] + "webp"
            object = resource.Object(test_bucket.name, webp)
            object.put(Body = bytes.getvalue(), ACL = 'public-read')

for bucket in resource.buckets.all():
    print(bucket.name)
    for i in bucket.objects.all():
        print(i.key)
        if (i.key[-3:] == "pdf"):
            file = s3.get_object(Bucket=bucket.name, Key=i.key)["Body"].read()
            pdf_stream = io.BytesIO(file)
            pdf_document = pymupdf.open(stream=pdf_stream, filetype="pdf")

            try:
                 page = pdf_document[0]
            except:
                 continue
            pix = page.get_pixmap()
            img_data = pix.tobytes("png")

            image = Image.open(io.BytesIO(img_data))
            #print(image.size)
            size = (10000, 120)
            image.thumbnail(size)
            #image.show()
            #print(type(image))

            bytes = BytesIO()
            image.save(bytes, format="webp")
            webp = i.key[:-3] + "webp"
            object = resource.Object(bucket.name, webp)
            object.put(Body = bytes.getvalue(),ACL='public-read')

#print(type(pdf_file))
#reader = PdfReader(BytesIO(pdf_file))

#print(f"Text: {reader.pages[0].extract_text()}")