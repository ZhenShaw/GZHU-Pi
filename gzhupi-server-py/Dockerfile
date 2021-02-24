# FROM aliyunfc/runtime-python3.6:1.7.7
FROM python:3.8

EXPOSE 5000

WORKDIR /code
ADD . /code

RUN pip install -r requirements.txt --target=./

ENTRYPOINT [ "" ]

CMD python app.py
