FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY actions.py ./actions/
EXPOSE 5055
CMD ["python", "-m", "rasa_sdk", "--actions", "actions"]
