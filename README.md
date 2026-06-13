# dos-29-diplomas-projects
## Site Checker

### Описание
Site Checker — простое веб-приложение на Go, которое позволяет отслеживать доступность сайтов (HTTP/HTTPS) и измерять время их отклика. В приложении есть веб-интерфейс для добавления сайтов, просмотра их статуса (работает/не работает) и анализа времени ответа через удобные дашборды. Для хранения данных используется база PostgreSQL.

### Функциональность
- Добавление сайтов для мониторинга через веб-форму
- Периодическая проверка доступности сайтов и времени отклика
- Дашборд с цветными индикаторами: зеленый — сайт доступен, красный — недоступен
- Дашборд с таблицей времени ответа сайтов

## Архитектура

Проект включает:
- GitHub Actions — CI/CD пайплайн для автоматического тестирования, сборки и деплоя
- Terraform — создание инфраструктуры в Yandex Cloud (ВМ, сеть, статический IP)
- Ansible — настройка сервера (Docker, Nginx, SSL)
- Docker — контейнеризация приложения и PostgreSQL
- Nginx + Let's Encrypt — HTTPS и безопасность

## Технологии

Backend:
- Go 1.21
- PostgreSQL 15

Инфраструктура:
- Terraform
- Ansible
- Yandex Cloud

CI/CD:
- GitHub Actions
- Yandex Container Registry

Контейнеризация:
- Docker
- Docker Compose

Безопасность:
- Nginx
- Certbot (SSL)
- Security Groups


## Развёртывание

1. Создание инфраструктуры:
   cd infra/terraform
   terraform init
   terraform apply

2. Настройка сервера:
   cd infra/ansible
   ansible-playbook -i inventory.ini playbook.yml

3. Деплой происходит автоматически при push в main

## CI/CD

При push в main:
1. Тесты и линтер
2. Сборка Docker-образа
3. Публикация в Yandex Container Registry
4. Деплой на сервер
---

Стек: Go, PostgreSQL, Docker, Terraform, Ansible, GitHub Actions, Yandex Cloud, Nginx, Let's Encrypt
