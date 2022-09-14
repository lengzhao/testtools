# testtools

easy test

## grpc client

1. 可以简单的通过json文件，就能够测试grpc服务的功能
2. 可以通过proto文件，自动生成测试模版（不含数据的json文件），补充少量数据就能够成为测试用例

### 如何用

1. 安装：go install github.com/lengzhao/testtools/cmd/grpc_testtool@latest
2. 查看帮助：grpc_testtool -h
   1. 如果命令失败，请确认是否将\$GOPATH/bin添加到\$PATH里面

    ```bash
    $grpc_testtool -h
    Usage of grpc_testtool:
    -addr string
            the grpc server address (default "localhost:50051")
    -gen string
            testcase path, new testcase with null value
    -import string
            import path of proto, split with ',' (default "./protos")
    -proto string
            proto path (default "./protos")
    -testcase string
            testcase path(include json files) (default "./testcase")
    ```

3. 将proto文件放到文件夹protos里面
4. 生成测试模板：grpc_testtool -gen ./testcase
   1. 它将自动根据不同的service创建不同的子文件夹
   2. 然后根据不同的method，创建不同的json文件
   3. 每个文件对应一个接口的测试参数
   4. 文件夹的名字和文件的名字都可以修改
5. 填充测试用例的数据

    ```json
    {
    "name": "helloworld.Greeter.SayHello",
    "service": "helloworld.Greeter",
    "method": "helloworld.Greeter.SayHello",
    "headers": [],
    "error_code": 0,
    "error": "",
    "request": {
        "name": "aa"
    },
    "response": {
        "message": "Hello aa"
    }
    }
    ```

   1. 样例数据如上
   2. name可以根据自己的需要修改，尽量不重复，方便确认执行的是哪个用例
   3. 自己可以添加其他字段，如description，描述用例的测试场景
   4. 如果有需要指定grpc header，使用[]string，每一项对应一个header
      1. 每个header用":"分隔为key和value
   5. request是请求需要的传的数据
   6. response是希望得到的数据，它将跟实际得到的数据进行比较
      1. 它要求完全匹配，否则失败
      2. 如果接受任何的响应，则设置为"response": "*"
   7. 如果是测试异常场景，则设置error和error_code
      1. 如果只是希望失败，但不限制具体的错误信息，则设置为error_code为非零值，"error": "*"
   8. 可以添加/复制/修改测试用例

6. 启动grpc服务，如greetee_server，假设其端口为5555
7. 执行测试：grpc_testtool -addr 127.0.0.1:5555 -testcase ./testcase -import ./protos -proto ./protos

## grpc server

1. 可以简单的通过json文件，就能够模拟简单的grpc服务
2. 可以通过proto文件，自动生成测试模版（不含数据的json文件），补充少量数据就能够成为模拟数据

### 如何使用

1. 安装：go install github.com/lengzhao/testtools/cmd/dynamic_grpc_server@latest
2. 查看帮助：dynamic_grpc_server -h
   1. 如果命令失败，请确认是否将\$GOPATH/bin添加到\$PATH里面

    ```bash
    % dynamic_grpc_server -h
    Usage of dynamic_grpc_server:
    -gen string
            testcase path, new testcase with null value
    -import string
            import path of proto, split with ',' (default "./protos")
    -port int
            The server port (default 50051)
    -proto string
            proto path (default "./protos")
    -testcase string
            testcase path(include json files) (default "./testcase")
    ```

3. 将proto文件放到文件夹protos里面
4. 生成模拟数据的模板：dynamic_grpc_server -gen ./testcase
   1. 它将自动根据不同的service创建不同的子文件夹
   2. 然后根据不同的method，创建不同的json文件
   3. 每个文件对应一个接口的模拟参数
   4. 文件夹的名字和文件的名字都可以修改
   5. 可以自己复制多个文件，填充不同数据，模拟不同场景
5. 填充测试用例的数据

    ```json
    {
    "name": "helloworld.Greeter.SayHello",
    "service": "helloworld.Greeter",
    "method": "helloworld.Greeter.SayHello",
    "headers": [],
    "error_code": 0,
    "error": "",
    "request": {
        "name": "aa"
    },
    "response": {
        "message": "Hello aa"
    }
    }
    ```

   1. 样例数据如上
   2. name可以根据自己的需要修改，尽量不重复，方便确认执行的是哪个用例
   3. 自己可以添加其他字段，如description，描述用例的模拟场景
   4. request是要匹配的请求数据
      1. 可以设置默认的，匹配任何数据，设置为"request": "*"
      2. 如果没有设置默认数据，又没有匹配的request，则返回失败
   5. response是希望得到的数据，它将是client要收到的数据
   6. 如果要返回异常，则设置error_code和error
   7. 可以添加/复制/修改测试用例

6. 启动grpc服务：dynamic_grpc_server
