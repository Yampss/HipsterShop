using System;
using System.Linq;
using System.Threading.Tasks;
using Grpc.Core;
using Google.Protobuf;
using Npgsql;

namespace cartservice.cartstore
{
    public class PostgresCartStore : ICartStore
    {
        private readonly string _connectionString;

        public PostgresCartStore(string connectionString)
        {
            _connectionString = connectionString;
        }

        public async Task AddItemAsync(string userId, string productId, int quantity)
        {
            Console.WriteLine($"AddItemAsync called with userId={userId}, productId={productId}, quantity={quantity}");

            try
            {
                using var conn = new NpgsqlConnection(_connectionString);
                await conn.OpenAsync();

                Hipstershop.Cart cart = new Hipstershop.Cart { UserId = userId };
                
                using (var cmd = new NpgsqlCommand("SELECT cart_data FROM carts WHERE user_id = @userId", conn))
                {
                    cmd.Parameters.AddWithValue("userId", userId);
                    using var reader = await cmd.ExecuteReaderAsync();
                    if (await reader.ReadAsync())
                    {
                        var bytes = (byte[])reader["cart_data"];
                        cart = Hipstershop.Cart.Parser.ParseFrom(bytes);
                    }
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

                using (var cmd = new NpgsqlCommand(@"
                    INSERT INTO carts (user_id, cart_data) VALUES (@userId, @cartData)
                    ON CONFLICT (user_id) DO UPDATE SET cart_data = EXCLUDED.cart_data", conn))
                {
                    cmd.Parameters.AddWithValue("userId", userId);
                    cmd.Parameters.AddWithValue("cartData", cart.ToByteArray());
                    await cmd.ExecuteNonQueryAsync();
                }
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
                using var conn = new NpgsqlConnection(_connectionString);
                await conn.OpenAsync();
                using var cmd = new NpgsqlCommand("DELETE FROM carts WHERE user_id = @userId", conn);
                cmd.Parameters.AddWithValue("userId", userId);
                await cmd.ExecuteNonQueryAsync();
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
                using var conn = new NpgsqlConnection(_connectionString);
                await conn.OpenAsync();
                using var cmd = new NpgsqlCommand("SELECT cart_data FROM carts WHERE user_id = @userId", conn);
                cmd.Parameters.AddWithValue("userId", userId);
                using var reader = await cmd.ExecuteReaderAsync();
                if (await reader.ReadAsync())
                {
                    var bytes = (byte[])reader["cart_data"];
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
                using var conn = new NpgsqlConnection(_connectionString);
                conn.Open();
                return true;
            }
            catch (Exception)
            {
                return false;
            }
        }
    }
}
