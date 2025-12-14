# WB API Connection Check
https://dev.wildberries.ru/en/openapi/api-information/#tag/WB-API-Connection-Check/paths/~1ping/get
https://common-api.wildberries.ru/ping

## Method description
Checks:

Whether the request successfully reaches the WB API.
The validity of the authorization token and request URL.
Whether the token category matches the service.
This method is not intended to check the availability of WB services
Each service has its own version of the method depending on the domain:

## info

Category	Request URL
Content	https://content-api.wildberries.ru/ping
https://content-api-sandbox.wildberries.ru/ping
Analytics	https://seller-analytics-api.wildberries.ru/ping
Prices and Discounts	https://discounts-prices-api.wildberries.ru/ping
https://discounts-prices-api-sandbox.wildberries.ru/ping
Marketplace	https://marketplace-api.wildberries.ru/ping
Statistics	https://statistics-api.wildberries.ru/ping
https://statistics-api-sandbox.wildberries.ru/ping
Promotion	https://advert-api.wildberries.ru/ping
https://advert-api-sandbox.wildberries.ru/ping
Feedbacks and Questions	https://feedbacks-api.wildberries.ru/ping
https://feedbacks-api-sandbox.wildberries.ru/ping
Buyers Chat	https://buyer-chat-api.wildberries.ru/ping
Supplies	https://supplies-api.wildberries.ru/ping
Buyers Returns	https://returns-api.wildberries.ru/ping
Documents	https://documents-api.wildberries.ru/ping
Finance	https://finance-api.wildberries.ru/ping
Tariffs, News, Seller Information	https://common-api.wildberries.ru/ping
Seller User Management	https://user-management-api.wildberries.ru/ping
A maximum of 3 requests every 30 seconds. If you try to use this method programmatically, the method will be temporarily blocked. The rate limit applies individually to each instance of the method on each host

## Response samples

Content type: application/json

Copy
{
"TS": "2024-08-16T11:19:05+03:00",
"Status": "OK"
}

