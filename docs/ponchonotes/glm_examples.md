GLM 4.6
basic call
curl -X POST "https://api.z.ai/api/paas/v4/chat/completions" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer your-api-key" \
-d '{
"model": "glm-4.6",
"messages": [
{
"role": "user",
"content": "As a marketing expert, please create an attractive slogan for my product."
},
{
"role": "assistant",
"content": "Sure, to craft a compelling slogan, please tell me more about your product."
},
{
"role": "user",
"content": "Z.AI Open Platform"
}
],
"thinking": {
"type": "enabled"
},
"max_tokens": 4096,
"temperature": 1.0
}'

Streaming Call

curl -X POST "https://api.z.ai/api/paas/v4/chat/completions" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer your-api-key" \
-d '{
"model": "glm-4.6",
"messages": [
{
"role": "user",
"content": "As a marketing expert, please create an attractive slogan for my product."
},
{
"role": "assistant",
"content": "Sure, to craft a compelling slogan, please tell me more about your product."
},
{
"role": "user",
"content": "Z.AI Open Platform"
}
],
"thinking": {
"type": "enabled"
},
"stream": true,
"max_tokens": 4096,
"temperature": 1.0
}'

GLM 4.6 Vision

basic call
curl -X POST \
    https://api.z.ai/api/paas/v4/chat/completions \
    -H "Authorization: Bearer your-api-key" \
    -H "Content-Type: application/json" \
    -d '{
        "model": "glm-4.6v",
        "messages": [
            {
                "role": "user",
                "content": [
                    {
                        "type": "image_url",
                        "image_url": {
                            "url": "https://cloudcovert-1305175928.cos.ap-guangzhou.myqcloud.com/%E5%9B%BE%E7%89%87grounding.PNG"
                        }
                    },
                    {
                        "type": "text",
                        "text": "Where is the second bottle of beer from the right on the table?  Provide coordinates in [[xmin,ymin,xmax,ymax]] format"
                    }
                ]
            }
        ],
        "thinking": {
            "type":"enabled"
        }
    }'

=======
streaming call example (vision)
curl -X POST \
    https://api.z.ai/api/paas/v4/chat/completions \
    -H "Authorization: Bearer your-api-key" \
    -H "Content-Type: application/json" \
    -d '{
        "model": "glm-4.6v",
        "messages": [
            {
                "role": "user",
                "content": [
                    {
                        "type": "image_url",
                        "image_url": {
                            "url": "https://cloudcovert-1305175928.cos.ap-guangzhou.myqcloud.com/%E5%9B%BE%E7%89%87grounding.PNG"
                        }
                    },
                    {
                        "type": "text",
                        "text": "Where is the second bottle of beer from the right on the table?  Provide coordinates in [[xmin,ymin,xmax,ymax]] format"
                    }
                ]
            }
        ],
        "thinking": {
            "type":"enabled"
        },
        "stream": true
    }'