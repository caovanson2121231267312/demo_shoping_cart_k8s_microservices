FROM python:3.11-slim
WORKDIR /app

COPY actions/requirements.txt ./requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

COPY bot_engine.py chat_data.py ./
COPY actions ./actions

ENV PYTHONPATH=/app
EXPOSE 5055
CMD ["python", "-m", "rasa_sdk", "--actions", "actions.actions"]
