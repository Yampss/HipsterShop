
using System;
using System.Threading.Tasks;
using Grpc.Core;
using Grpc.Health.V1;
using static Grpc.Health.V1.Health;
using cartservice.cartstore;

namespace cartservice.services
{
    internal class HealthCheckService : HealthBase
    {
        private ICartStore _cartStore { get; }

        public HealthCheckService (ICartStore cartStore) 
        {
            _cartStore = cartStore;
        }

        public override Task<HealthCheckResponse> Check(HealthCheckRequest request, ServerCallContext context)
        {
            Console.WriteLine ("Checking CartService Health");
            return Task.FromResult(new HealthCheckResponse {
                Status = _cartStore.Ping() ? HealthCheckResponse.Types.ServingStatus.Serving : HealthCheckResponse.Types.ServingStatus.NotServing
            });
        }
    }
}
