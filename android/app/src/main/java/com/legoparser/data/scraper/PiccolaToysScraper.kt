package com.legoparser.data.scraper

import org.jsoup.Jsoup
import org.jsoup.nodes.Document

class PiccolaToysScraper {
    suspend fun scrape(): List<ScrapedProduct> {
        val products = mutableListOf<ScrapedProduct>()
        val detailUrls = mutableListOf<String>()

        // Collect product URLs from search pages
        var page = 1
        while (true) {
            val url = if (page == 1) "https://piccolatoys.ge/?s=lego&post_type=product"
                else "https://piccolatoys.ge/page/$page/?s=lego&post_type=product"
            val doc = fetchDoc(url) ?: break

            val links = doc.select(".wd-product .product-image-link")
            if (links.isEmpty()) break
            links.forEach { el -> el.attr("href").takeIf { it.isNotBlank() }?.let { detailUrls.add(it) } }
            if (doc.select("a.next.page-numbers").isEmpty()) break
            page++
        }

        // Scrape each product detail
        for (detailUrl in detailUrls) {
            val doc = fetchDoc(detailUrl) ?: continue
            val name = doc.select("h1.product_title").text().trim()
            if (name.isBlank()) continue

            val sku = doc.select("span.sku").text().trim()
            val delPrice = doc.select("p.price del .amount").text().let { parsePrice(it) }
            val insPrice = doc.select("p.price ins .amount").text().let { parsePrice(it) }
            val singlePrice = doc.select("p.price .amount").first()?.text()?.let { parsePrice(it) }

            var original = delPrice ?: singlePrice
            var discounted: Double? = null
            var pct: Double? = null
            if (delPrice != null && insPrice != null) {
                original = delPrice; discounted = insPrice
                pct = calcDiscountPercent(delPrice, insPrice)
            }

            products.add(ScrapedProduct(
                invoiceCode = sku.ifBlank { extractLegoID(name) },
                nameKa = name, originalPrice = original,
                discountPercent = pct, discountedPrice = discounted,
                sourceUrl = detailUrl
            ))
        }

        return products
    }

    private fun fetchDoc(url: String): Document? = try {
        Jsoup.connect(url)
            .userAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
            .timeout(30000).get()
    } catch (_: Exception) { null }
}
