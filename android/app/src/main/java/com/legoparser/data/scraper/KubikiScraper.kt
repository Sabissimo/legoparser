package com.legoparser.data.scraper

import android.content.Context

class KubikiScraper(private val context: Context) {
    suspend fun scrape(): List<ScrapedProduct> {
        val products = mutableListOf<ScrapedProduct>()

        for (page in 1..40) {
            val url = "https://kubiki.ge/shop/page/$page/?count=36"
            val doc = try {
                WebViewScraper.fetchRenderedHtml(context, url, waitMs = 5000)
            } catch (_: Exception) { break }

            var found = 0
            for (el in doc.select("li.product, .product-item, .product")) {
                val name = el.select(".product-title a, h3 a, h2 a, .woocommerce-loop-product__title").first()?.text()?.trim() ?: continue
                if (name.isBlank()) continue
                // Kubiki is LEGO-only store, no filter needed

                val delPrice = el.select("del .woocommerce-Price-amount, del .amount").first()?.text()?.let { parsePrice(it) }
                val insPrice = el.select("ins .woocommerce-Price-amount, ins .amount").first()?.text()?.let { parsePrice(it) }
                val singlePrice = el.select(".price .woocommerce-Price-amount, .price .amount, .price").first()?.text()?.let { parsePrice(it) }

                var original = delPrice ?: singlePrice
                var discounted: Double? = null
                var pct: Double? = null
                if (delPrice != null && insPrice != null) {
                    original = delPrice; discounted = insPrice
                    pct = calcDiscountPercent(delPrice, insPrice)
                }

                val productUrl = el.select("a").first()?.attr("href")
                val imgUrl = el.select("img").first()?.let { it.attr("src").ifBlank { it.attr("data-src") } }

                products.add(ScrapedProduct(
                    invoiceCode = extractLegoID(name), nameKa = name,
                    originalPrice = original, discountPercent = pct, discountedPrice = discounted,
                    imageUrl = imgUrl, sourceUrl = productUrl
                ))
                found++
            }
            if (found == 0) break
        }
        return products
    }
}
