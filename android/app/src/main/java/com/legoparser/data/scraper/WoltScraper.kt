package com.legoparser.data.scraper

import android.content.Context

class WoltScraper(private val context: Context, private val source: String, private val stores: List<String>) {

    companion object {
        val xsStores = listOf("xs-toys-galleria-tbilisi1", "xs-toys-city-mall", "lego-galleria-tbilisi", "xs-toys-tbilisi-mall")
        val pepelaStores = listOf("pepela-vake-park", "pepela-city-mall", "pepela-vake", "pepela-marjanishvili", "pepela-saburtalo", "pepela-aghmashenebeli")
    }

    suspend fun scrape(): List<ScrapedProduct> {
        val products = mutableListOf<ScrapedProduct>()

        for (slug in stores) {
            val url = "https://wolt.com/en/geo/tbilisi/venue/$slug/items/lego-3"
            val doc = try {
                WebViewScraper.fetchRenderedHtml(context, url, waitMs = 5000)
            } catch (_: Exception) { continue }

            for (el in doc.select("[data-test-id=ItemCard]")) {
                val name = el.select("[data-test-id=ImageCentricProductCard.Title], h3").first()?.text()?.trim() ?: continue
                if (!isLegoProduct(name)) continue

                var priceText = el.select("[aria-label*=Price]").first()?.text() ?: el.text()
                priceText = priceText.replace("GEL", "").trim()
                val price = parsePrice(priceText)

                val imgUrl = el.select("[data-test-id=ImageCentricProductCard.ProductImage], img").first()?.attr("src")
                val productUrl = el.select("[data-test-id=CardLinkButton], a").first()?.attr("href")?.let {
                    if (it.startsWith("http")) it else "https://wolt.com$it"
                }

                products.add(ScrapedProduct(
                    invoiceCode = extractLegoID(name), nameKa = name,
                    originalPrice = price, imageUrl = imgUrl, sourceUrl = productUrl
                ))
            }
        }
        return products
    }
}
