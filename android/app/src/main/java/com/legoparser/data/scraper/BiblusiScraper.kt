package com.legoparser.data.scraper

import com.google.gson.Gson
import com.google.gson.JsonObject
import okhttp3.OkHttpClient
import okhttp3.Request
import java.util.concurrent.TimeUnit

class BiblusiScraper(private val source: String, private val categoryId: Int) {
    private val client = OkHttpClient.Builder().readTimeout(30, TimeUnit.SECONDS).build()
    private val gson = Gson()

    suspend fun scrape(): List<ScrapedProduct> {
        val products = mutableListOf<ScrapedProduct>()
        var page = 1

        while (true) {
            val url = "https://apiv1.biblusi.ge/api/book?category_id=$categoryId&per_page=20&page=$page&category=1&author=1"
            val body = fetch(url) ?: break
            val json = gson.fromJson(body, JsonObject::class.java)
            val data = json.getAsJsonArray("data") ?: break

            for (element in data) {
                val book = element.asJsonObject
                val name = book.get("name")?.asString ?: continue
                if (!isLegoProduct(name)) continue

                val bookId = book.get("id")?.asString ?: ""
                val price = book.get("p")?.asDouble
                val legoId = extractLegoID(name)

                products.add(ScrapedProduct(
                    mcode = bookId,
                    invoiceCode = legoId,
                    nameKa = name,
                    originalPrice = price,
                    sourceUrl = "https://biblusi.ge/products/$bookId"
                ))
            }

            val lastPage = json.get("last_page")?.asInt ?: 1
            if (page >= lastPage) break
            page++
        }

        return products
    }

    private fun fetch(url: String): String? {
        val request = Request.Builder().url(url)
            .header("User-Agent", "Mozilla/5.0")
            .header("Accept", "application/json")
            .build()
        return try {
            client.newCall(request).execute().use { it.body?.string() }
        } catch (_: Exception) { null }
    }
}
