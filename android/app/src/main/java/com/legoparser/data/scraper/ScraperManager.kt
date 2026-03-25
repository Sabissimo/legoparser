package com.legoparser.data.scraper

import android.content.Context
import com.legoparser.data.db.AppDatabase
import com.legoparser.data.db.PriceEntity
import com.legoparser.data.db.ProductEntity
import com.legoparser.data.db.ScrapeRunEntity
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import kotlinx.coroutines.withContext

class ScraperManager(private val context: Context) {
    private val db = AppDatabase.get(context)
    private val mutex = Mutex()

    val allSources = listOf(
        "biblusi_xs", "biblusi_pepela", "wishlist", "piccolatoys",
        "kubiki", "wolt_xs", "wolt_pepela", "glovo_xs", "glovo_pepela"
    )

    suspend fun runAll() {
        for (source in allSources) {
            runSource(source)
        }
    }

    suspend fun runSource(source: String) = mutex.withLock {
        val run = ScrapeRunEntity(source = source)
        val runId = db.scrapeRunDao().insert(run)
        run.id == runId

        try {
            val products = withContext(Dispatchers.IO) {
                when (source) {
                    "biblusi_xs" -> BiblusiScraper("biblusi_xs", 456).scrape()
                    "biblusi_pepela" -> BiblusiScraper("biblusi_pepela", 474).scrape()
                    "wishlist" -> WishlistScraper(context).scrape()
                    "piccolatoys" -> PiccolaToysScraper().scrape()
                    "kubiki" -> KubikiScraper(context).scrape()
                    "wolt_xs" -> WoltScraper(context, "wolt_xs", WoltScraper.xsStores).scrape()
                    "wolt_pepela" -> WoltScraper(context, "wolt_pepela", WoltScraper.pepelaStores).scrape()
                    "glovo_xs" -> GlovoScraper(context, "glovo_xs", GlovoScraper.xsUrl).scrape()
                    "glovo_pepela" -> GlovoScraper(context, "glovo_pepela", GlovoScraper.pepelaUrl).scrape()
                    else -> emptyList()
                }
            }

            var saved = 0
            for (sp in products) {
                val productId = upsertProduct(sp)
                db.priceDao().insert(PriceEntity(
                    productId = productId,
                    source = source,
                    originalPrice = sp.originalPrice,
                    discountPercent = sp.discountPercent,
                    discountedPrice = sp.discountedPrice,
                    inStock = sp.inStock,
                    sourceUrl = sp.sourceUrl
                ))
                saved++
            }

            db.scrapeRunDao().update(ScrapeRunEntity(
                id = runId, source = source, status = "completed",
                productsFound = products.size, productsSaved = saved,
                startedAt = run.startedAt, completedAt = System.currentTimeMillis()
            ))
        } catch (e: Exception) {
            db.scrapeRunDao().update(ScrapeRunEntity(
                id = runId, source = source, status = "failed",
                errorsCount = 1, errorLog = e.message,
                startedAt = run.startedAt, completedAt = System.currentTimeMillis()
            ))
        }
    }

    private suspend fun upsertProduct(sp: ScrapedProduct): Long {
        val dao = db.productDao()

        // Match by invoice_code → barcode → name
        sp.invoiceCode?.takeIf { it.isNotBlank() }?.let { code ->
            dao.findByInvoiceCode(code)?.let { id ->
                dao.update(id, sp.mcode, sp.barcode, sp.invoiceCode, sp.imageUrl)
                return id
            }
        }
        sp.barcode?.takeIf { it.isNotBlank() }?.let { barcode ->
            dao.findByBarcode(barcode)?.let { id ->
                dao.update(id, sp.mcode, sp.barcode, sp.invoiceCode, sp.imageUrl)
                return id
            }
        }
        dao.findByName(sp.nameKa)?.let { id ->
            dao.update(id, sp.mcode, sp.barcode, sp.invoiceCode, sp.imageUrl)
            return id
        }

        return dao.insert(ProductEntity(
            mcode = sp.mcode, barcode = sp.barcode, invoiceCode = sp.invoiceCode,
            nameKa = sp.nameKa, nameEn = sp.nameEn, imageUrl = sp.imageUrl
        ))
    }
}
