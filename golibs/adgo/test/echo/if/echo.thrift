//namespace go echo

struct EchoRequest {
    1: string message
}

struct EchoResponse {
    1: string message
}

exception Exception {
    1: string reason
}

service EchoService {
    EchoResponse echo(1: EchoRequest request) throws (1: Exception e)
}
