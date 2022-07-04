# Speculator

A library for reconstructing OpenAPI specification from traffic of HTTP transactions.

### Limitations

1. Doesn't populate media type encoding.
2. OAuth2 flow type can't be known, we are using authorizationCode for now.