# Stress Test - Golang Expert

Para buildar a aplicação use o comando docker abaixo:

```
docker build -t stress-test .
```

Para rodar a aplicação, use um dos comandos abaixo:

```
go run main.go --url=https://google.com --requests=100 --concurrency=20
```

ou

```
docker run stress-test --url=http://google.com --concurrency=20 --requests=300
```