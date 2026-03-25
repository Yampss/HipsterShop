#!/bin/bash
docker build -t cazzzzz/adservice-hipster:latest ./src/adservice
docker push cazzzzz/adservice-hipster:latest
docker build -t cazzzzz/cartservice-hipster:latest ./src/cartservice/src
docker push cazzzzz/cartservice-hipster:latest
docker build -t cazzzzz/checkoutservice-hipster:latest ./src/checkoutservice
docker push cazzzzz/checkoutservice-hipster:latest
docker build -t cazzzzz/currencyservice-hipster:latest ./src/currencyservice
docker push cazzzzz/currencyservice-hipster:latest
docker build -t cazzzzz/emailservice-hipster:latest ./src/emailservice
docker push cazzzzz/emailservice-hipster:latest
docker build -t cazzzzz/frontend-hipster:latest ./src/frontend
docker push cazzzzz/frontend-hipster:latest
docker build -t cazzzzz/loadgenerator-hipster:latest ./src/loadgenerator
docker push cazzzzz/loadgenerator-hipster:latest
docker build -t cazzzzz/paymentservice-hipster:latest ./src/paymentservice
docker push cazzzzz/paymentservice-hipster:latest
docker build -t cazzzzz/productcatalogservice-hipster:latest ./src/productcatalogservice
docker push cazzzzz/productcatalogservice-hipster:latest
docker build -t cazzzzz/recommendationservice-hipster:latest ./src/recommendationservice
docker push cazzzzz/recommendationservice-hipster:latest
docker build -t cazzzzz/shippingservice-hipster:latest ./src/shippingservice
docker push cazzzzz/shippingservice-hipster:latest
