package com.legoparser.ui.screens

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.ArrowForward
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.legoparser.viewmodel.ProductsViewModel

@Composable
fun ProductsScreen(vm: ProductsViewModel = viewModel()) {
    val state by vm.state.collectAsStateWithLifecycle()
    var searchInput by remember { mutableStateOf("") }

    Column(modifier = Modifier.fillMaxSize().padding(16.dp)) {
        OutlinedTextField(
            value = searchInput, onValueChange = { searchInput = it },
            modifier = Modifier.fillMaxWidth(),
            placeholder = { Text("Search name, LEGO ID, MCode...") },
            leadingIcon = { Icon(Icons.Default.Search, null) },
            singleLine = true,
            trailingIcon = {
                TextButton(onClick = { vm.setSearch(searchInput) }) { Text("Go") }
            }
        )
        Spacer(Modifier.height(8.dp))

        Row(horizontalArrangement = Arrangement.spacedBy(4.dp)) {
            listOf("name_ka" to "Name", "invoice_code" to "LEGO ID", "mcode" to "MCode").forEach { (key, label) ->
                val active = state.sortBy == key
                FilterChip(
                    selected = active, onClick = { vm.toggleSort(key) },
                    label = { Text("$label${if (active) (if (state.sortOrder == "asc") " ↑" else " ↓") else ""}", fontSize = 11.sp) },
                    modifier = Modifier.height(30.dp)
                )
            }
        }
        Spacer(Modifier.height(4.dp))
        Text("${state.total} products", style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
        Spacer(Modifier.height(4.dp))

        if (state.isLoading) {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { CircularProgressIndicator() }
        } else {
            LazyColumn(Modifier.weight(1f)) {
                items(state.products) { p ->
                    Card(Modifier.fillMaxWidth().padding(vertical = 3.dp)) {
                        Column(Modifier.padding(12.dp)) {
                            Text(p.nameKa, fontWeight = FontWeight.Medium, maxLines = 2, overflow = TextOverflow.Ellipsis)
                            Spacer(Modifier.height(4.dp))
                            Row(horizontalArrangement = Arrangement.spacedBy(16.dp)) {
                                p.invoiceCode?.let { Label("LEGO ID", it) }
                                p.mcode?.let { Label("MCode", it) }
                                p.barcode?.let { Label("Barcode", it) }
                            }
                            p.sources?.takeIf { it.isNotBlank() }?.let { src ->
                                Spacer(Modifier.height(6.dp))
                                Row(horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                                    src.split(",").forEach { s ->
                                        SuggestionChip(onClick = {}, label = { Text(s.trim().replace("_", " "), fontSize = 9.sp) }, modifier = Modifier.height(22.dp))
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

@Composable
private fun Label(label: String, value: String) {
    Column {
        Text(label, fontSize = 10.sp, color = MaterialTheme.colorScheme.onSurfaceVariant)
        Text(value, fontSize = 12.sp, fontFamily = FontFamily.Monospace)
    }
}
