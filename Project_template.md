## Задание 1

**Диаграмма контейнеров (Containers)**

[C4_Container.pumpl](./C4_containers.pumpl)

## Задание 2

### 1. Proxy

[API Gateway](./src/microservices/proxy/)

### 2. Kafka

[tests](./tests/scrinshots/tests.png)

[topic movie-events](./tests/scrinshots/movie_events.png)

[topic user-events](./tests/scrinshots/user-events.png)

[topic payment-events](./tests/scrinshots/payment-events.png)

[consumers](./tests/scrinshots/consumers.png)

[log events service](./tests/scrinshots/log%20event%20service.png)

## Задание 3

[Actions GitHub](https://github.com/Karinka-malinka/swa-project-sprint-2/actions)

### Proxy в Kubernetes

[get movies](./tests/scrinshots/get%20movies.png)

[log events service](./tests/scrinshots/log%20events.png)

[test kubernetes](./tests/scrinshots/test%20kuber.mov)

## Задание 4

kubectl delete all --all -n cinemaabyss
kubectl delete  namespace cinemaabyss
helm install cinemaabyss src/kubernetes/helm --namespace cinemaabyss --create-namespace
kubectl -n cinemaabyss get pod

[helm](./tests/scrinshots/helm.png)

# Задание 5

[Статистика работы circuit breaker'а](./tests/scrinshots/statistics%201.png)

[Статистика работы circuit breaker'а](./tests/scrinshots/statistics%202.png)