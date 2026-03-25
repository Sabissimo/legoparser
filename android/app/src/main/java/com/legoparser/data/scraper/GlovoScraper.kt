package com.legoparser.data.scraper

import android.content.Context

class GlovoScraper(private val context: Context, private val source: String, private val url: String) {

    companion object {
        const val xsUrl = "https://glovoapp.com/en/ge/tbilisi/stores/xs-toys-tbi?content=lego-konstruqtorebi-c.21175679"
        const val pepelaUrl = "https://glovoapp.com/en/ge/tbilisi/stores/pepela-tbi?content=lego-sc.21169575%2Flego-konstruqtorebi-c.21169580"
    }

    suspend fun scrape(): List<ScrapedProduct> {
        val doc = try {
            WebViewScraper.fetchRenderedHtml(context, url, waitMs = 5000)
        } catch (_: Exception) { return emptyList() }

        val products = mutableListOf<ScrapedProduct>()

        for (el in doc.select("[class*=ItemTile_itemTile]")) {
            val name = el.select("[class*=ItemTile_title], h3").first()?.text()?.trim() ?: continue

            val priceText = el.select("[class*=ItemTile_discountedPrice], [class*=price]").first()?.text() ?: ""
            val price = parsePrice(priceText)

            val imgUrl = el.select("[class*=ItemTile_image] img, img").first()?.attr("src")

            products.add(ScrapedProduct(
                invoiceCode = extractLegoID(name), nameKa = name,
                originalPrice = price, imageUrl = imgUrl, sourceUrl = url
            ))
        }
        return products
    }
}
