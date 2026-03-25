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
import com.legoparser.data.db.PriceComparisonRow
import com.legoparser.viewmodel.ComparisonViewModel

private data class PCol(val label: String, val get: (PriceComparisonRow) -> Double?)
private val cols = listOf(
    PCol("Bibl XS") { it.biblusi_xs }, PCol("Bibl Pep") { it.biblusi_pepela },
    PCol("Wolt XS") { it.wolt_xs }, PCol("Wolt Pep") { it.wolt_pepela },
    PCol("Glov XS") { it.glovo_xs }, PCol("Glov Pep") { it.glovo_pepela },
    PCol("Wish") { it.wishlist }, PCol("Piccola") { it.piccolatoys }, PCol("Kubiki") { it.kubiki }
)

@Composable
fun ComparisonScreen(vm: ComparisonViewModel = viewModel()) {
    val state by vm.state.collectAsStateWithLifecycle()
    var searchInput by remember { mutableStateOf("") }

    Column(Modifier.fillMaxSize().padding(16.dp)) {
        OutlinedTextField(
            value = searchInput, onValueChange = { searchInput = it },
            modifier = Modifier.fillMaxWidth(), placeholder = { Text("Search...") },
            leadingIcon = { Icon(Icons.Default.Search, null) }, singleLine = true,
            trailingIcon = { TextButton(onClick = { vm.setSearch(searchInput) }) { Text("Go") } }
        )
        Spacer(Modifier.height(8.dp))

        if (state.isLoading) {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { CircularProgressIndicator() }
        } else {
            LazyColumn(Modifier.weight(1f)) {
                items(state.items) { item ->
                    val prices = cols.map { it.get(item) }
                    val valid = prices.filterNotNull()
                    val min = valid.minOrNull()
                    val max = valid.maxOrNull()

                    Card(Modifier.fillMaxWidth().padding(vertical = 3.dp)) {
                        Column(Modifier.padding(12.dp)) {
                            Text(item.nameKa, fontWeight = FontWeight.Medium, maxLines = 2, overflow = TextOverflow.Ellipsis)
                            item.invoiceCode?.let {
                                Text("LEGO $it", fontSize = 11.sp, color = MaterialTheme.colorScheme.primary, fontFamily = FontFamily.Monospace)
                            }
                            Spacer(Modifier.height(8.dp))
                            Row(Modifier.horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                                prices.forEachIndexed { i, price ->
                                    val bg = when {
                                        price == null || valid.size < 2 -> Color.Transparent
                                        price == min -> Color(0xFFDCFCE7)
                                        price == max -> Color(0xFFFEE2E2)
                                        else -> Color.Transparent
                                    }
                                    val tc = when {
                                        price == null -> MaterialTheme.colorScheme.onSurfaceVariant
                                        valid.size < 2 -> MaterialTheme.colorScheme.onSurface
                                        price == min -> Color(0xFF166534)
                                        price == max -> Color(0xFF991B1B)
                                        else -> MaterialTheme.colorScheme.onSurface
                                    }
                                    Surface(color = bg, shape = MaterialTheme.shapes.small, modifier = Modifier.width(58.dp)) {
                                        Column(Modifier.padding(3.dp), horizontalAlignment = Alignment.CenterHorizontally) {
                                            Text(cols[i].label, fontSize = 8.sp, color = MaterialTheme.colorScheme.onSurfaceVariant)
                                            Text(
                                                if (price != null) "%.1f".format(price) else "-",
                                                fontSize = 10.sp, fontFamily = FontFamily.Monospace,
                                                fontWeight = if (price == min && valid.size >= 2) FontWeight.Bold else FontWeight.Normal,
                                                color = tc, textAlign = TextAlign.Center
                                            )
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
            if (state.totalPages > 1) {
                Row(Modifier.fillMaxWidth().padding(top = 4.dp), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
                    Text("${state.page} / ${state.totalPages}", style = MaterialTheme.typography.bodySmall)
                    Row {
                        IconButton(onClick = { vm.setPage(state.page - 1) }, enabled = state.page > 1) { Icon(Icons.Default.ArrowBack, "Prev") }
                        IconButton(onClick = { vm.setPage(state.page + 1) }, enabled = state.page < state.totalPages) { Icon(Icons.Default.ArrowForward, "Next") }
                    }
                }
            }
        }
    }
}
