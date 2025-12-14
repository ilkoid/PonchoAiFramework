# Wildberries API Research

## Overview
This document contains comprehensive information about Wildberries API gathered through web research for integration with Poncho Tools.

## API Documentation and Resources

### Official Documentation
- **Main API Portal**: [dev.wildberries.ru](https://dev.wildberries.ru/en?from=seller_landing)
- **API Information**: [dev.wildberries.ru/en/openapi/api-information](https://dev.wildberries.ru/en/openapi/api-information)
- **Developer Documentation**: [dev.wildberries.ru/en](https://dev.wildberries.ru/en?from=seller_landing)
- **FAQ Section**: [dev.wildberries.ru/en/faq](https://dev.wildberries.ru/en/faq)
- **Sandbox Environment**: [dev.wildberries.ru/en/openapi-other/sandbox-environment](https://dev.wildberries.ru/en/openapi-other/sandbox-environment)

### Key API Modules
The Wildberries API supports **11 main modules**:
1. **Products** - Product management and information
2. **Orders** - Order management (FBS/FBW)
3. **Finances** - Financial data and transactions
4. **Analytics** - Sales analytics and statistics
5. **Reports** - Various reporting functions
6. **Communications** - Customer communication
7. **Promotions** - Marketing and promotions
8. **Tariffs** - Pricing and tariff information
9. **Documents** - Document management
10. **Digital** - Wildberries Digital operations
11. **Content** - Product content and categories

## Authentication and Authorization

### API Token Requirements
- **Registration**: Required in seller personal account
- **Token Creation**: Created in store settings
- **Token Validity**: 180 days (must be regenerated after expiration)
- **Authorization Method**: Token-based authorization for most requests

### New Authorization Methods
- **Service Secrets**: New approach for third-party services
- **Business Solutions Catalog**: For service integrations
- **OAuth Authentication**: Available for certain integrations

### Rate Limiting
- **Algorithm**: Token bucket algorithm for load distribution
- **Request Limits**: Specific limits for different API categories
- **Content Category**: Special limits for content-related methods
- **Sales Funnel**: Maximum limit parameters for statistics (updated December 8)

## Key Endpoints and Features

### Product Management
- **Product Cards List**: [dev.wildberries.ru/openapi/work-with-products](https://dev.wildberries.ru/openapi/work-with-products)
- **HS-codes**: Available by category name with filtering
- **Product Information**: Comprehensive product data management

### Analytics and Data
- **Analytics Endpoint**: [dev.wildberries.ru/openapi/analytics](https://dev.wildberries.ru/openapi/analytics)
- **Sales Statistics**: Detailed sales funnel analytics
- **Performance Metrics**: Real-time and statistical information


## Technical Specifications

### API Protocol
- **Protocol**: HTTP REST API
- **Data Format**: JSON
- **Authentication**: Bearer token
- **Rate Limiting**: Token bucket algorithm

### Development Tools
- **Mock Servers**: For development and testing
- **Sandbox Environment**: Safe testing environment
- **OpenAPI Specifications**: Available for some endpoints

## Integration Considerations

### Relevant API Modules for Product Descriptions
1. **Products API**: For product schema and category information
2. **Content API**: For product content requirements and validation
3. **Analytics API**: For market trend analysis
4. **Categories API**: For product classification and schema requirements

### Implementation Strategy
1. **Schema Retrieval**: Use Products API to get category-specific schemas
2. **Validation**: Integrate with Content API for validation rules
3. **Category Detection**: Use API to determine product categories
4. **Market Analysis**: Leverage Analytics API for trend data

### Rate Limit Management
- **Request Optimization**: Batch requests where possible
- **Caching Strategy**: Implement local caching for schemas
- **Error Handling**: Graceful degradation for rate limits
- **Retry Logic**: Exponential backoff for failed requests

## Development Resources

### Documentation Links
- [Main Developer Portal](https://dev.wildberries.ru/en?from=seller_landing)
- [API Information](https://dev.wildberries.ru/en/openapi/api-information)
- [FAQ Section](https://dev.wildberries.ru/en/faq)
- [Release Notes](https://dev.wildberries.ru/en/release-notes)

### Community and Support
- **GitHub Topics**: [wildberries-api](https://github.com/topics/wildberries-api)
- **Case Studies**: Multiple integration examples
- **Developer Community**: Active developer support


## Security and Compliance

### Data Protection
- **API Token Security**: Secure token management
- **Data Encryption**: HTTPS/TLS encryption
- **Access Control**: Proper authorization handling

### Compliance Requirements
- **Rate Limiting**: Respect API limits
- **Data Usage**: Follow Wildberries data policies
- **Privacy Protection**: Customer data protection
