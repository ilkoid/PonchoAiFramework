# Brands

https://dev.wildberries.ru/en/openapi/work-with-products/#tag/Categories-Subjects-and-Characteristics/paths/~1api~1content~1v1~1brands/get
https://content-api.wildberries.ru/api/content/v1/brands

## Method description
The method returns list of brands by subject ID.

Request limit per one seller's account:
Period	Limit	Interval	Burst
1 second	1 request	1 second	5 requests

## query Parameters
subjectId
required
integer
Example: subjectId=1234
Subject ID

next	
integer
Example: next=1234
Pagination parameter. Use the next value from the response to get the next data batch

## Response samples 200

{
  "brands": [
    {
      "id": 9007199254,
      "logoUrl": "string",
      "name": "Brand"
    }
  ],
  "next": 1212,
  "total": 344534
}
