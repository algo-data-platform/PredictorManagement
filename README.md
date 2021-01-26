# Content_service / Predictor_platform

## Introduction
- Content_service 是模型拉取及配置通知agent，跟预测服务同机部署，同时可以单独部署为一致性灰度模型验证服务。
- Predictor_platform 为预测服务的开放平台，用来对模型，服务，机器，权重，线程池，模型负责人等进行简单的可视化配置，查看服务模型加载状态，机器状态。同时支持包括集群调整，自动压测，自动权重，一键降级等控制功能。

## Get Started - Clone the repo
### 1.clone PredictorManagement repo
```sh
$ git clone https://github.com/algo-data-platform/PredictorManagement.git
```
### 2.start predictor service
```sh
$ git clone https://github.com/algo-data-platform/Predictorservice.git
$ cd PredictorService/runtime
$ sh ./start_predictor.sh
```

## Get Started - Build Predictor_platform
### 1. init test db
(assuming you are at the repo base dir: `PredictorManagement/`)
```sh
$ cd predictor_platform
$ sh ./init_test_db.sh
```
### 2. build
(assuming you are at the repo base dir: `PredictorManagement/`)
```sh
$ cd predictor_platform
$ sh ./build.sh
```

## Get Started - Run Predictor_platform
### 1. start
(assuming you are at the repo base dir: `PredictorManagement/`)
```sh
$ cd predictor_platform/server
$ sh ./run.sh
```
it should print out a message with an url to see the server status, such as:
> login in predictor platform http://local_host:10008/
```
username: admin
password: admin
```

## Get Started - Build Content_service
### 1. build
(assuming you are at the repo base dir: `PredictorManagement/`)
```sh
$ cd content_serivce
$ sh ./build.sh
```
## Get Started - Run Content_service
### 1. start content service
(assuming you are at the repo base dir: `PredictorManagement/`)
```sh
$ cd content_serivce/
$ sh ./run.sh
```
it should print out a message with an url to see the server status, such as:
> check predictor status on http://local_host:10018/server/status

after 30s content_service will notify predictor to load models, then predictor service 
should print out a message with an url to see the model status, such as:
> check model status on http://local_host:10048/get_service_model_info

至此，通过predictor platform控制，content_service拉取模型并通知Predictor service，实现了动态的上线模型及注册服务。