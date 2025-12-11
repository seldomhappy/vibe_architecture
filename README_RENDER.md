# Деплой на Render.com

## Быстрый старт

### 1. Подготовка

1. Создайте аккаунт на [Render.com](https://render.com)
2. Подключите ваш GitHub репозиторий

### 2. Деплой через Dashboard

1. Перейдите в Render Dashboard
2. Нажмите "New +" → "Blueprint"
3. Подключите репозиторий `seldomhappy/vibe_architecture`
4. Render автоматически обнаружит `render.yaml`
5. Нажмите "Apply"

Render создаст:
- ✅ PostgreSQL database (free tier)
- ✅ Web service с Docker
- ✅ Автоматическую связь между сервисами

### 3. Ручной деплой (альтернатива)

#### 3.1 Создать PostgreSQL

1. New + → PostgreSQL
2. Name: `vibe-db`
3. Database: `vibe_architecture`
4. User: `vibe_user`
5. Plan: Free
6. Create Database

#### 3.2 Создать Web Service

1. New + → Web Service
2. Подключить репозиторий
3. Name: `vibe-architecture`
4. Environment: Docker
5. Region: Oregon (или ближайший)
6. Branch: main
7. Plan: Free

#### 3.3 Настроить Environment Variables

Добавить переменные (или использовать значения из `render.yaml`):

```
APP_NAME=vibe-architecture
APP_ENVIRONMENT=production
DB_HOST=[из PostgreSQL Internal Database URL]
DB_PORT=5432
DB_USER=vibe_user
DB_PASSWORD=[из PostgreSQL]
DB_NAME=vibe_architecture
DB_SSL_MODE=require
KAFKA_ENABLED=false
TRACING_ENABLED=false
METRICS_ENABLED=true
```

#### 3.4 Deploy

Нажмите "Create Web Service"

### 4. Проверка деплоя

После успешного деплоя:

```bash
# Health check
curl https://your-app.onrender.com/health

# Создать задачу
curl -X POST https://your-app.onrender.com/tasks \
  -H "Content-Type: application/json" \
  -d '{"name":"Test task","priority":"high"}'

# Получить задачи
curl https://your-app.onrender.com/tasks

# Метрики
curl https://your-app.onrender.com:9090/metrics
```

### 5. Автоматические деплои

Render автоматически деплоит при push в `main` ветку.

### 6. Логи

Просмотр логов:
1. Dashboard → vibe-architecture → Logs
2. Или через CLI: `render logs -s vibe-architecture`

### 7. Мониторинг

Render предоставляет:
- CPU/Memory usage
- Request metrics
- Crash reports

### 8. Ограничения Free Tier

⚠️ **Важно:**
- Бесплатный сервис "засыпает" после 15 минут неактивности
- Первый запрос после сна займет ~30 секунд
- 750 часов/месяц бесплатно
- PostgreSQL: 1GB storage, 97 часов/месяц активности

Для production используйте платный план ($7/месяц).

### 9. Troubleshooting

**Проблема:** Сервис не стартует
**Решение:** Проверьте логи, убедитесь что миграции прошли успешно

**Проблема:** Database connection failed
**Решение:** Проверьте что DB_SSL_MODE=require

**Проблема:** Долгий cold start
**Решение:** Это норма для free tier. Используйте платный план или UptimeRobot для keep-alive

### 10. Включение Kafka (опционально)

Если нужна Kafka в будущем:

1. Используйте [CloudKarafka](https://www.cloudkarafka.com/) (free tier)
2. Или [Upstash Kafka](https://upstash.com/) (serverless)
3. Обновите env variables:
   ```
   KAFKA_ENABLED=true
   KAFKA_BROKERS=kafka-url:9092
   ```

### 11. Включение Tracing (опционально)

Для трассировки используйте:
- [Honeycomb](https://www.honeycomb.io/) (free tier)
- [Grafana Cloud](https://grafana.com/products/cloud/) (free tier)

```
TRACING_ENABLED=true
JAEGER_ENDPOINT=https://your-jaeger-endpoint/api/traces
```

## Готово! 🎉

Ваш микросервис запущен на Render.com!
