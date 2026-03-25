package com.legoparser.api

import com.legoparser.BuildConfig
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import retrofit2.http.*
import java.util.concurrent.TimeUnit

interface LegoParserApi {
    @GET("products")
    suspend fun getProducts(
        @Query("page") page: Int = 1,
        @Query("per_page") perPage: Int = 20,
        @Query("search") search: String? = null,
        @Query("sort_by") sortBy: String = "name_ka",
        @Query("sort_order") sortOrder: String = "asc"
    ): PaginatedResponse<Product>

    @GET("products/comparison")
    suspend fun getComparison(
        @Query("page") page: Int = 1,
        @Query("per_page") perPage: Int = 20,
        @Query("search") search: String? = null,
        @Query("sort_by") sortBy: String = "name_ka",
        @Query("sort_order") sortOrder: String = "asc"
    ): PaginatedResponse<PriceComparison>

    @GET("scrape/runs")
    suspend fun getScrapeRuns(): List<ScrapeRun>

    @GET("scrape/status")
    suspend fun getScrapeStatus(): ScrapeStatus

    @POST("scrape/run")
    suspend fun startScrape(@Body request: ScrapeRequest): Map<String, String>
}

object ApiClient {
    private val client = OkHttpClient.Builder()
        .connectTimeout(30, TimeUnit.SECONDS)
        .readTimeout(60, TimeUnit.SECONDS)
        .addInterceptor(HttpLoggingInterceptor().apply {
            level = HttpLoggingInterceptor.Level.BASIC
        })
        .build()

    val api: LegoParserApi = Retrofit.Builder()
        .baseUrl(BuildConfig.API_BASE_URL + "/")
        .client(client)
        .addConverterFactory(GsonConverterFactory.create())
        .build()
        .create(LegoParserApi::class.java)
}
