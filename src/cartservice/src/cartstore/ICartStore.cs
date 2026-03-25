
using System.Threading.Tasks;

namespace cartservice.cartstore
{
    public interface ICartStore
    {
        Task AddItemAsync(string userId, string productId, int quantity);
        Task EmptyCartAsync(string userId);
        Task<Hipstershop.Cart> GetCartAsync(string userId);
        bool Ping();
    }
}
