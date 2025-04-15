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
print(type(pdf_file))
reader = PdfReader(BytesIO(pdf_file))

print(f"Text: {reader.pages[0].extract_text()}")

# page = reader.pages[0]
# pix = page.get_pixmap()
# img_path = "temp_pdf_page.png"
# pix.save(img_path)

pdf_stream = io.BytesIO(pdf_file)
pdf_document = pymupdf.open(stream=pdf_stream, filetype="pdf")

page = pdf_document[0]
pix = page.get_pixmap()
img_data = pix.tobytes("png")

image = Image.open(io.BytesIO(img_data))
print(image.size)
size = (300, 300)
image.thumbnail(size)
image.show()
print(type(image))

bytes = BytesIO()
image.save(bytes, format="webp")

resource = boto3.resource('s3')
object = resource.Object("bep-json-test-bucket", 'files/0010.webp')
object.put(Body = bytes.getvalue())


pdf_document.close()

# img = Image.open(img_path)
# img.show()

# os.remove(img_path)
