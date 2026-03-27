using System;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Diagnostics.HealthChecks;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Http;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Diagnostics.HealthChecks;
using Microsoft.Extensions.Hosting;
using cartservice.cartstore;
using cartservice.services;
using Microsoft.Extensions.Caching.StackExchangeRedis;
using OpenTelemetry.Resources;
using OpenTelemetry.Trace;

namespace cartservice
{
    public class Startup
    {
        public Startup(IConfiguration configuration)
        {
            Configuration = configuration;
        }

        public IConfiguration Configuration { get; }
        
        public void ConfigureServices(IServiceCollection services)
        {
            string mongoConnectionString = Configuration["MONGO_CONNECTION_STRING"];
            string redisAddress = Configuration["REDIS_ADDR"];

            if (string.IsNullOrEmpty(mongoConnectionString))
            {
                throw new InvalidOperationException("MONGO_CONNECTION_STRING environment variable is required.");
            }

            // MongoDB is always the primary store
            var mongoStore = new MongoCartStore(mongoConnectionString);
            Console.WriteLine("MongoDB cart store created (primary)");

            if (!string.IsNullOrEmpty(redisAddress))
            {
                // Redis available — use as write-through cache in front of MongoDB
                Console.WriteLine($"Redis cache enabled at {redisAddress}");
                services.AddStackExchangeRedisCache(options =>
                {
                    options.Configuration = redisAddress;
                });
                services.AddSingleton<MongoCartStore>(mongoStore);
                services.AddSingleton<ICartStore, RedisCartStore>();
            }
            else
            {
                // No Redis — use MongoDB directly
                Console.WriteLine("No Redis configured, using MongoDB directly");
                services.AddSingleton<ICartStore>(mongoStore);
            }

            services.AddGrpc();

            services.AddOpenTelemetry()
                .WithTracing(builder => builder
                    .AddAspNetCoreInstrumentation()
                    .AddGrpcClientInstrumentation()
                    .AddOtlpExporter());
        }

        public void Configure(IApplicationBuilder app, IWebHostEnvironment env)
        {
            if (env.IsDevelopment())
            {
                app.UseDeveloperExceptionPage();
            }

            app.UseRouting();

            app.UseEndpoints(endpoints =>
            {
                endpoints.MapGrpcService<CartService>();
                endpoints.MapGrpcService<cartservice.services.HealthCheckService>();

                endpoints.MapGet("/", async context =>
                {
                    await context.Response.WriteAsync("Communication with gRPC endpoints must be made through a gRPC client. To learn how to create a client, visit: https://go.microsoft.com/fwlink/?linkid=2086909");
                });
            });
        }
    }
}
