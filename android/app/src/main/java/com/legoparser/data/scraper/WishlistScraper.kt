package com.legoparser.data.scraper

import android.content.Context

class WishlistScraper(private val context: Context) {
    private val baseUrl = "https://wishlist.ge/%E1%83%91%E1%83%90%E1%83%95%E1%83%A8%E1%83%95%E1%83%97%E1%83%90-%E1%83%A1%E1%83%90%E1%83%9B%E1%83%A7%E1%83%90%E1%83%A0%E1%83%9D/%E1%83%A1%E1%83%90%E1%83%97%E1%83%90%E1%83%9B%E1%83%90%E1%83%A8%E1%83%9D%E1%83%94%E1%83%91%E1%83%98/lego-%E1%83%9A%E1%83%94%E1%83%92%E1%83%9D-ka"

    suspend fun scrape(): List<ScrapedProduct> {
        val products = mutableListOf<ScrapedProduct>()

        // Find last page
        val maxPage = findLastPage()

        for (page in 1..maxPage) {
            val url = "$baseUrl/page-$page/?items_per_page=192"
            val doc = try {
                WebViewScraper.fetchRenderedHtml(context, url, waitMs = 4000)
            } catch (_: Exception) { break }

            var found = 0
            for (el in doc.select(".ty-grid-list__item")) {
                val name = el.select("a.product-title").text().ifBlank {
                    el.select(".ty-grid-list__item-name a").text()
                }.trim()
                if (name.isBlank()) continue

                val productId = el.select("input[name*=product_data][name*=product_id]").attr("value")
                val legoId = extractLegoIDAnywhere(name)

                val invoiceCode: String?
                val mcode: String?
                if (legoId != null) {
                    invoiceCode = legoId
                    mcode = productId.ifBlank { null }
                } else {
                    invoiceCode = productId.ifBlank { null }
                    mcode = null
                }

                val oldPrice = el.select(".ty-strike .ty-list-price").text().let { parsePrice(it) }
                val newPrice = el.select(".ty-price .ty-price-num").text().let { parsePrice(it) }

                var originalPrice = oldPrice ?: newPrice
                var discountedPrice: Double? = null
                var discountPct: Double? = null
                if (oldPrice != null && newPrice != null && newPrice < oldPrice) {
                    originalPrice = oldPrice
                    discountedPrice = newPrice
                    discountPct = calcDiscountPercent(oldPrice, newPrice)
                }

                val productUrl = el.select("a.product-title").attr("href")
                val imgUrl = el.select("img.ty-pict").attr("src").ifBlank {
                    el.select("img.ty-pict").attr("data-src")
                }

                products.add(ScrapedProduct(
                    mcode = mcode, invoiceCode = invoiceCode, nameKa = name,
                    originalPrice = originalPrice, discountPercent = discountPct,
                    discountedPrice = discountedPrice, imageUrl = imgUrl.ifBlank { null },
                    sourceUrl = productUrl.ifBlank { null }
                ))
                found++
            }
            if (found == 0) break
        }

        return products
    }

    private suspend fun findLastPage(): Int {
        for (page in 20 downTo 1) {
            val url = "$baseUrl/page-$page/?items_per_page=192"
            val doc = try {
                WebViewScraper.fetchRenderedHtml(context, url, waitMs = 4000)
            } catch (_: Exception) { continue }
            val bodyText = doc.body().text()
            if ("ვერ იქნა მოძიებული" in bodyText || "404" in bodyText) continue
            if (doc.select(".ty-grid-list__item").isNotEmpty()) return page
        }
        return 1
    }
}
