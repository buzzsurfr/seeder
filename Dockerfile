FROM python:3-alpine

RUN pip3 install boto3

COPY main.py /main.py
CMD ["python3", "/main.py"]
