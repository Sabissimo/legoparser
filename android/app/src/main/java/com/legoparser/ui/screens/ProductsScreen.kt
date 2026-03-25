package com.legoparser.ui.screens

import androidx.compose.foundation.clickable
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

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ProductsScreen(vm: ProductsViewModel = viewModel()) {
    val state by vm.state.collectAsStateWithLifecycle()
    var searchInput by remember { mutableStateOf("") }

    Column(modifier = Modifier.fillMaxSize().padding(16.dp)) {
        // Search bar
        OutlinedTextField(
            value = searchInput,
            onValueChange = { searchInput = it },
            modifier = Modifier.fillMaxWidth(),
            placeholder = { Text("Search by name, LEGO ID, MCode...") },
            leadingIcon = { Icon(Icons.Default.Search, contentDescription = null) },
            singleLine = true,
            trailingIcon = {
                if (searchInput.isNotBlank()) {
                    TextButton(onClick = { vm.setSearch(searchInput) }) {
                        Text("Search")
                    }
                }
            }
        )

        Spacer(modifier = Modifier.height(8.dp))

        // Header
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                "${state.total} products",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            Row {
                SortChip("Name", "name_ka", state.sortBy, state.sortOrder) { vm.toggleSort("name_ka") }
                SortChip("LEGO ID", "invoice_code", state.sortBy, state.sortOrder) { vm.toggleSort("invoice_code") }
                SortChip("MCode", "mcode", state.sortBy, state.sortOrder) { vm.toggleSort("mcode") }
            }
        }

        Spacer(modifier = Modifier.height(8.dp))

        if (state.isLoading) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator()
            }
        } else if (state.error != null) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Text("Error: ${state.error}", color = MaterialTheme.colorScheme.error)
            }
        } else {
            LazyColumn(modifier = Modifier.weight(1f)) {
                items(state.products) { product ->
                    Card(
                        modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp),
                        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)
                    ) {
                        Column(modifier = Modifier.padding(12.dp)) {
                            Text(
                                product.name_ka,
                                style = MaterialTheme.typography.bodyMedium,
                                fontWeight = FontWeight.Medium,
                                maxLines = 2,
                                overflow = TextOverflow.Ellipsis
                            )
                            Spacer(modifier = Modifier.height(4.dp))
                            Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                                if (product.invoice_code != null) {
                                    LabelValue("LEGO ID", product.invoice_code)
                                }
                                if (product.mcode != null) {
                                    LabelValue("MCode", product.mcode)
                                }
                                if (product.barcode != null) {
                                    LabelValue("Barcode", product.barcode)
                                }
                            }
                            if (product.sources.isNotEmpty()) {
                                Spacer(modifier = Modifier.height(6.dp))
                                Row(horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                                    product.sources.forEach { src ->
                                        SuggestionChip(
                                            onClick = {},
                                            label = {
                                                Text(
                                                    src.replace("_", " "),
                                                    fontSize = 10.sp
                                                )
                                            },
                                            modifier = Modifier.height(24.dp)
                                        )
                                    }
                                }
                            }
                        }
                    }
                }
            }

            // Pagination
            if (state.totalPages > 1) {
                Row(
                    modifier = Modifier.fillMaxWidth().padding(top = 8.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text("Page ${state.page} / ${state.totalPages}", style = MaterialTheme.typography.bodySmall)
                    Row {
                        IconButton(
                            onClick = { vm.setPage(state.page - 1) },
                            enabled = state.page > 1
                        ) { Icon(Icons.Default.ArrowBack, "Previous") }
                        IconButton(
                            onClick = { vm.setPage(state.page + 1) },
                            enabled = state.page < state.totalPages
                        ) { Icon(Icons.Default.ArrowForward, "Next") }
                    }
                }
            }
        }
    }
}

@Composable
private fun LabelValue(label: String, value: String) {
    Column {
        Text(label, fontSize = 10.sp, color = MaterialTheme.colorScheme.onSurfaceVariant)
        Text(value, fontSize = 12.sp, fontFamily = FontFamily.Monospace)
    }
}

@Composable
private fun SortChip(label: String, key: String, currentSort: String, currentOrder: String, onClick: () -> Unit) {
    val isActive = currentSort == key
    val arrow = if (isActive) (if (currentOrder == "asc") " ↑" else " ↓") else ""
    FilterChip(
        selected = isActive,
        onClick = onClick,
        label = { Text("$label$arrow", fontSize = 11.sp) },
        modifier = Modifier.padding(horizontal = 2.dp).height(30.dp)
    )
}
