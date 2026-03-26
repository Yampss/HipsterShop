using System;
using System.Linq;
using System.Threading.Tasks;
using Grpc.Core;
using Google.Protobuf;
using MongoDB.Driver;
using MongoDB.Bson;

namespace cartservice.cartstore
{
    public class MongoCartStore : ICartStore
    {
        private readonly IMongoCollection<BsonDocument> _cartsCollection;

        public MongoCartStore(string connectionString)
        {
            var client = new MongoClient(connectionString);
            var database = client.GetDatabase("shopdb");
            _cartsCollection = database.GetCollection<BsonDocument>("carts");
        }

        public async Task AddItemAsync(string userId, string productId, int quantity)
        {
            Console.WriteLine($"AddItemAsync called with userId={userId}, productId={productId}, quantity={quantity}");

            try
            {
                Hipstershop.Cart cart = new Hipstershop.Cart { UserId = userId };

                var filter = Builders<BsonDocument>.Filter.Eq("user_id", userId);
                var existingDoc = await _cartsCollection.Find(filter).FirstOrDefaultAsync();

                if (existingDoc != null && existingDoc.Contains("cart_data"))
                {
                    var bytes = existingDoc["cart_data"].AsByteArray;
                    cart = Hipstershop.Cart.Parser.ParseFrom(bytes);
                }

                var existingItem = cart.Items.SingleOrDefault(i => i.ProductId == productId);
                if (existingItem == null)
                {
                    cart.Items.Add(new Hipstershop.CartItem { ProductId = productId, Quantity = quantity });
                }
                else
                {
                    existingItem.Quantity += quantity;
                }

                await _cartsCollection.ReplaceOneAsync(
                    filter,
                    new BsonDocument
                    {
                        { "user_id", userId },
                        { "cart_data", new BsonBinaryData(cart.ToByteArray()) }
                    },
                    new ReplaceOptions { IsUpsert = true });
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
                var filter = Builders<BsonDocument>.Filter.Eq("user_id", userId);
                await _cartsCollection.DeleteOneAsync(filter);
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
                var filter = Builders<BsonDocument>.Filter.Eq("user_id", userId);
                var doc = await _cartsCollection.Find(filter).FirstOrDefaultAsync();

                if (doc != null && doc.Contains("cart_data"))
                {
                    var bytes = doc["cart_data"].AsByteArray;
                    return Hipstershop.Cart.Parser.ParseFrom(bytes);
                }

                return new Hipstershop.Cart { UserId = userId };
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
                _cartsCollection.Database.Client
                    .GetDatabase("admin")
                    .RunCommand<BsonDocument>(new BsonDocument("ping", 1));
                return true;
            }
            catch (Exception)
            {
                return false;
            }
        }
    }
}
