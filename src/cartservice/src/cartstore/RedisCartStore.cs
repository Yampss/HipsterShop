
using System;
using System.Linq;
using System.Threading.Tasks;
using Grpc.Core;
using Microsoft.Extensions.Caching.Distributed;
using Google.Protobuf;

namespace cartservice.cartstore
{
    public class RedisCartStore : ICartStore
    {
        private readonly IDistributedCache _cache;
        private readonly MongoCartStore _mongoStore;

        public RedisCartStore(IDistributedCache cache, MongoCartStore mongoStore)
        {
            _cache = cache;
            _mongoStore = mongoStore;
        }

        public async Task AddItemAsync(string userId, string productId, int quantity)
        {
            Console.WriteLine($"AddItemAsync called with userId={userId}, productId={productId}, quantity={quantity}");

            try
            {
                // Write-through: persist to MongoDB first
                await _mongoStore.AddItemAsync(userId, productId, quantity);

                // Then update the Redis cache
                var cart = await _mongoStore.GetCartAsync(userId);
                await _cache.SetAsync(userId, cart.ToByteArray());
            }
            catch (Exception ex)
            {
                throw new RpcException(new Status(StatusCode.FailedPrecondition, $"Can't access cart storage. {ex}"));
            }
        }

        public async Task EmptyCartAsync(string userId)
        {
            Console.WriteLine($"EmptyCartAsync called with userId={userId}");

            try
            {
                // Delete from MongoDB
                await _mongoStore.EmptyCartAsync(userId);

                // Invalidate Redis cache
                await _cache.RemoveAsync(userId);
            }
            catch (Exception ex)
            {
                throw new RpcException(new Status(StatusCode.FailedPrecondition, $"Can't access cart storage. {ex}"));
            }
        }

        public async Task<Hipstershop.Cart> GetCartAsync(string userId)
        {
            Console.WriteLine($"GetCartAsync called with userId={userId}");

            try
            {
                // Try Redis cache first
                var value = await _cache.GetAsync(userId);
                if (value != null)
                {
                    return Hipstershop.Cart.Parser.ParseFrom(value);
                }

                // Cache miss — load from MongoDB and cache it
                var cart = await _mongoStore.GetCartAsync(userId);
                if (cart.Items.Count > 0)
                {
                    await _cache.SetAsync(userId, cart.ToByteArray());
                }

                return cart;
            }
            catch (Exception ex)
            {
                throw new RpcException(new Status(StatusCode.FailedPrecondition, $"Can't access cart storage. {ex}"));
            }
        }

        public bool Ping()
        {
            try
            {
                return _mongoStore.Ping();
            }
            catch (Exception)
            {
                return false;
            }
        }
    }
}
