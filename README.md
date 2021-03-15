# Proxy Server
Проксирование запросов http, https. Repeater и сканнер уязвимости os command injection для get и post параметров.  

# Repeater usage
По дефолту на :8888
- GET /requests - список запросов
- GET /requests/{id} - отдельный запрос
- DELETE /requests - удаление всех запросов
- DELETE /requests/{id} - удаление конкретного запроса
- POST /requests/{id}/repeat - повторение запроса
- POST /requests/{id}/scan/cmd - сканирование запроса

# Author
Быховец Марк