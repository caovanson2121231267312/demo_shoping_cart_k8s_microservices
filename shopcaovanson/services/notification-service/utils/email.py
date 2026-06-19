import os
import smtplib
from email.mime.application import MIMEApplication
from email.mime.multipart import MIMEMultipart
from email.mime.text import MIMEText
from pathlib import Path
from typing import Any


def _env(primary: str, fallback: str, default: str = "") -> str:
    return os.getenv(primary) or os.getenv(fallback, default)


def _env_bool(primary: str, fallback: str, default: bool = True) -> bool:
    raw = _env(primary, fallback, "true" if default else "false")
    return raw.lower() in ("1", "true", "yes", "tls", "ssl")


class EmailSender:
    def __init__(self) -> None:
        self.host = _env("SMTP_HOST", "MAIL_HOST", "localhost")
        self.port = int(_env("SMTP_PORT", "MAIL_PORT", "587"))
        self.username = _env("SMTP_USERNAME", "MAIL_USERNAME", "")
        self.password = _env("SMTP_PASSWORD", "MAIL_PASSWORD", "")
        self.from_email = _env("SMTP_FROM", "MAIL_FROM", "noreply@shopcaovanson.xyz")
        self.use_tls = _env_bool("SMTP_USE_TLS", "MAIL_ENCRYPTION", True)

    def send(self, to_email: str, subject: str, html_body: str) -> None:
        msg = MIMEMultipart("alternative")
        msg["Subject"] = subject
        msg["From"] = self.from_email
        msg["To"] = to_email
        msg.attach(MIMEText(html_body, "html", "utf-8"))

        with smtplib.SMTP(self.host, self.port, timeout=30) as server:
            if self.use_tls:
                server.starttls()
            if self.username and self.password:
                server.login(self.username, self.password)
            server.sendmail(self.from_email, [to_email], msg.as_string())

    def send_template(self, to_email: str, subject: str, template, context: dict[str, Any]) -> None:
        html_body = template.render(**context)
        self.send(to_email, subject, html_body)

    def send_with_attachment(
        self,
        to_email: str,
        subject: str,
        html_body: str,
        attachment_path: Path,
        attachment_name: str,
    ) -> None:
        msg = MIMEMultipart("mixed")
        msg["Subject"] = subject
        msg["From"] = self.from_email
        msg["To"] = to_email

        alt = MIMEMultipart("alternative")
        alt.attach(MIMEText(html_body, "html", "utf-8"))
        msg.attach(alt)

        with open(attachment_path, "rb") as f:
            part = MIMEApplication(f.read(), Name=attachment_name)
        part["Content-Disposition"] = f'attachment; filename="{attachment_name}"'
        msg.attach(part)

        with smtplib.SMTP(self.host, self.port, timeout=30) as server:
            if self.use_tls:
                server.starttls()
            if self.username and self.password:
                server.login(self.username, self.password)
            server.sendmail(self.from_email, [to_email], msg.as_string())
