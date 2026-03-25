package com.legoparser.data.db

import androidx.room.*

@Entity(tableName = "products")
data class ProductEntity(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    val mcode: String? = null,
    val barcode: String? = null,
    @ColumnInfo(name = "invoice_code") val invoiceCode: String? = null,
    @ColumnInfo(name = "name_ka") val nameKa: String,
    @ColumnInfo(name = "name_en") val nameEn: String? = null,
    @ColumnInfo(name = "image_url") val imageUrl: String? = null,
    @ColumnInfo(name = "created_at") val createdAt: Long = System.currentTimeMillis(),
    @ColumnInfo(name = "updated_at") val updatedAt: Long = System.currentTimeMillis()
)

@Entity(
    tableName = "prices",
    foreignKeys = [ForeignKey(entity = ProductEntity::class, parentColumns = ["id"], childColumns = ["product_id"], onDelete = ForeignKey.CASCADE)],
    indices = [Index("product_id", "source")]
)
data class PriceEntity(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    @ColumnInfo(name = "product_id") val productId: Long,
    val source: String,
    @ColumnInfo(name = "original_price") val originalPrice: Double? = null,
    @ColumnInfo(name = "discount_percent") val discountPercent: Double? = null,
    @ColumnInfo(name = "discounted_price") val discountedPrice: Double? = null,
    @ColumnInfo(name = "in_stock") val inStock: Boolean = true,
    @ColumnInfo(name = "source_url") val sourceUrl: String? = null,
    @ColumnInfo(name = "scraped_at") val scrapedAt: Long = System.currentTimeMillis()
)

@Entity(tableName = "scrape_runs")
data class ScrapeRunEntity(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    val source: String,
    var status: String = "running",
    @ColumnInfo(name = "trigger_type") val triggerType: String = "manual",
    @ColumnInfo(name = "products_found") var productsFound: Int = 0,
    @ColumnInfo(name = "products_saved") var productsSaved: Int = 0,
    @ColumnInfo(name = "errors_count") var errorsCount: Int = 0,
    @ColumnInfo(name = "error_log") var errorLog: String? = null,
    @ColumnInfo(name = "started_at") val startedAt: Long = System.currentTimeMillis(),
    @ColumnInfo(name = "completed_at") var completedAt: Long? = null
)

data class ProductWithSources(
    val id: Long,
    val mcode: String?,
    val barcode: String?,
    @ColumnInfo(name = "invoice_code") val invoiceCode: String?,
    @ColumnInfo(name = "name_ka") val nameKa: String,
    @ColumnInfo(name = "name_en") val nameEn: String?,
    @ColumnInfo(name = "image_url") val imageUrl: String?,
    val sources: String? // comma-separated
)

data class PriceComparisonRow(
    @ColumnInfo(name = "product_id") val productId: Long,
    val mcode: String?,
    val barcode: String?,
    @ColumnInfo(name = "invoice_code") val invoiceCode: String?,
    @ColumnInfo(name = "name_ka") val nameKa: String,
    @ColumnInfo(name = "name_en") val nameEn: String?,
    val biblusi_xs: Double?,
    val biblusi_pepela: Double?,
    val wolt_xs: Double?,
    val wolt_pepela: Double?,
    val glovo_xs: Double?,
    val glovo_pepela: Double?,
    val wishlist: Double?,
    val piccolatoys: Double?,
    val kubiki: Double?
)
