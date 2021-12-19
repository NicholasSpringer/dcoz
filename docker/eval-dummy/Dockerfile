FROM python:3

ENV PYTHONUNBUFFERED=1

WORKDIR /app
COPY ./eval/dummy/requirements.txt .

RUN pip3 install -r requirements.txt

COPY ./eval/dummy .