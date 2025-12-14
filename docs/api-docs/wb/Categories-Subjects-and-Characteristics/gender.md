# Gender

https://dev.wildberries.ru/en/openapi/work-with-products/#tag/Categories-Subjects-and-Characteristics/paths/~1content~1v2~1directory~1kinds/get

https://content-api.wildberries.ru/content/v2/directory/kinds

## Method description
Provides values of gender characteristic.

Request limit per one seller's account for all methods in the Content category:
Period	Limit	Interval	Burst
1 minute	100 requests	600 milliseconds	5 requests

### Exceptions are the methods:

creating product cards
creating product cards with merge
editing product cards
getting failed product cards with errors

## query Parameters
locale	
string
Example: locale=en
Language for response of the subjectName and name fields:

ru — Russian
en — English
zh — Chinese
Not used in the sandbox

## Response samples 200:

{
  "data": [
    "Мужской"
  ],
  "error": false,
  "errorText": "",
  "additionalErrors": ""
}

