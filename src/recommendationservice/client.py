
import sys
import grpc
import demo_pb2
import demo_pb2_grpc

from logger import getJSONLogger
logger = getJSONLogger('recommendationservice-server')

if __name__ == "__main__":
    if len(sys.argv) > 1:
        port = sys.argv[1]
    else:
        port = "8080"

    channel = grpc.insecure_channel('localhost:'+port)
    stub = demo_pb2_grpc.RecommendationServiceStub(channel)
    request = demo_pb2.ListRecommendationsRequest(user_id="test", product_ids=["test"])
    response = stub.ListRecommendations(request)
    logger.info(response)
