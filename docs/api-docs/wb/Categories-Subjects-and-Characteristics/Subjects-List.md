# Subjects List

https://dev.wildberries.ru/en/openapi/work-with-products/#tag/Categories-Subjects-and-Characteristics/paths/~1content~1v2~1object~1all/get

https://content-api.wildberries.ru/content/v2/object/all

## Method description:
Returns the list of all available subjects, subjects parent categories and their IDs

Request limit per one seller's account for all methods in the Content category:
Period	Limit	Interval	Burst
1 minute	100 requests	600 milliseconds	5 requests
Exceptions are the methods:

creating product cards
creating product cards with merge
editing product cards
getting failed product cards with errors

## query Parameters:
locale	
string
Example: locale=en
Language for response of the name field:

ru — Russian
en — English
zh — Chinese
Not used in the sandbox

name	
string
Example: name=Socks
Search by item name (Socks), the search works by substring and can be conducted in any of the supported languages

limit	
integer
Default: 30
Example: limit=1000
Number of search results, maximum 1,000

offset	
integer
Default: 0
Example: offset=5000
How many results to skip. For example, with value 10, the response will start with the 11 element

parentID	
integer
Example: parentID=1000
Subject parent category ID

## Response samples 200:

{
  "data": [
    {
      "subjectID": 2560,
      "parentID": 479,
      "subjectName": "3D очки",
      "parentName": "Электроника"
    },
    {
      "subjectID": 1152,
      "parentID": 858,
      "subjectName": "3D-принтеры",
      "parentName": "Оргтехника"
    }
  ],
  "error": false,
  "errorText": "",
  "additionalErrors": null
}
