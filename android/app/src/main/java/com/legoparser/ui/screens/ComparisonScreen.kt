package com.legoparser.ui.screens

import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.ArrowForward
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.legoparser.api.PriceComparison
import com.legoparser.viewmodel.ComparisonViewModel

private data class PriceCol(val label: String, val key: String, val getter: (PriceComparison) -> Double?)

private val priceCols = listOf(
    PriceCol("Bibl XS", "biblusi_xs_price") { it.biblusi_xs_price },
    PriceCol("Bibl Pep", "biblusi_pepela_price") { it.biblusi_pepela_price },
    PriceCol("Wolt XS", "wolt_xs_price") { it.wolt_xs_price },
    PriceCol("Wolt Pep", "wolt_pepela_price") { it.wolt_pepela_price },
    PriceCol("Glov XS", "glovo_xs_price") { it.glovo_xs_price },
    PriceCol("Glov Pep", "glovo_pepela_price") { it.glovo_pepela_price },
    PriceCol("Wish", "wishlist_price") { it.wishlist_price },
    PriceCol("Piccola", "piccolatoys_price") { it.piccolatoys_price },
    PriceCol("Kubiki", "kubiki_price") { it.kubiki_price },
)

@Composable
fun ComparisonScreen(vm: ComparisonViewModel = viewModel()) {
    val state by vm.state.collectAsStateWithLifecycle()
    var searchInput by remember { mutableStateOf("") }

    Column(modifier = Modifier.fillMaxSize().padding(16.dp)) {
        OutlinedTextField(
            value = searchInput,
            onValueChange = { searchInput = it },
            modifier = Modifier.fillMaxWidth(),
            placeholder = { Text("Search by name, LEGO ID...") },
            leadingIcon = { Icon(Icons.Default.Search, contentDescription = null) },
            singleLine = true,
            trailingIcon = {
                if (searchInput.isNotBlank()) {
                    TextButton(onClick = { vm.setSearch(searchInput) }) { Text("Search") }
                }
            }
        )

        Spacer(modifier = Modifier.height(8.dp))

        if (state.isLoading) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator()
            }
        } else {
            LazyColumn(modifier = Modifier.weight(1f)) {
                items(state.items) { item ->
                    ComparisonCard(item, vm)
                }
            }

            if (state.totalPages > 1) {
                Row(
                    modifier = Modifier.fillMaxWidth().padding(top = 8.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text("Page ${state.page} / ${state.totalPages}", style = MaterialTheme.typography.bodySmall)
                    Row {
                        IconButton(onClick = { vm.setPage(state.page - 1) }, enabled = state.page > 1) {
                            Icon(Icons.Default.ArrowBack, "Previous")
                        }
                        IconButton(onClick = { vm.setPage(state.page + 1) }, enabled = state.page < state.totalPages) {
                            Icon(Icons.Default.ArrowForward, "Next")
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun ComparisonCard(item: PriceComparison, vm: ComparisonViewModel) {
    val prices = priceCols.map { it.getter(item) }
    val validPrices = prices.filterNotNull()
    val minPrice = validPrices.minOrNull()
    val maxPrice = validPrices.maxOrNull()

    Card(
        modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)
    ) {
        Column(modifier = Modifier.padding(12.dp)) {
            Text(
                item.name_ka,
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.Medium,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis
            )
            if (item.invoice_code != null) {
                Text(
                    "LEGO ${item.invoice_code}",
                    fontSize = 11.sp,
                    color = MaterialTheme.colorScheme.primary,
                    fontFamily = FontFamily.Monospace
                )
            }

            Spacer(modifier = Modifier.height(8.dp))

            Row(
                modifier = Modifier.horizontalScroll(rememberScrollState()),
                horizontalArrangement = Arrangement.spacedBy(6.dp)
            ) {
                priceCols.forEachIndexed { i, col ->
                    val price = prices[i]
                    val bgColor = when {
                        price == null -> Color.Transparent
                        validPrices.size < 2 -> Color.Transparent
                        price == minPrice -> Color(0xFFDCFCE7)
                        price == maxPrice -> Color(0xFFFEE2E2)
                        else -> Color.Transparent
                    }
                    val textColor = when {
                        price == null -> MaterialTheme.colorScheme.onSurfaceVariant
                        validPrices.size < 2 -> MaterialTheme.colorScheme.onSurface
                        price == minPrice -> Color(0xFF166534)
                        price == maxPrice -> Color(0xFF991B1B)
                        else -> MaterialTheme.colorScheme.onSurface
                    }

                    Surface(
                        color = bgColor,
                        shape = MaterialTheme.shapes.small,
                        modifier = Modifier.width(62.dp)
                    ) {
                        Column(
                            modifier = Modifier.padding(4.dp),
                            horizontalAlignment = Alignment.CenterHorizontally
                        ) {
                            Text(col.label, fontSize = 9.sp, color = MaterialTheme.colorScheme.onSurfaceVariant)
                            Text(
                                if (price != null) "%.2f".format(price) else "-",
                                fontSize = 11.sp,
                                fontFamily = FontFamily.Monospace,
                                fontWeight = if (price == minPrice && validPrices.size >= 2) FontWeight.Bold else FontWeight.Normal,
                                color = textColor,
                                textAlign = TextAlign.Center
                            )
                        }
                    }
                }
            }
        }
    }
}
