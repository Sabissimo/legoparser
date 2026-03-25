package com.legoparser.ui.screens

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.legoparser.viewmodel.ScraperViewModel

private val sources = listOf(
    "all" to "All Sources",
    "biblusi_xs" to "Biblusi XS",
    "biblusi_pepela" to "Biblusi Pepela",
    "wishlist" to "Wishlist",
    "piccolatoys" to "PiccolaToys",
    "kubiki" to "Kubiki",
    "wolt_xs" to "Wolt XS",
    "wolt_pepela" to "Wolt Pepela",
    "glovo_xs" to "Glovo XS",
    "glovo_pepela" to "Glovo Pepela",
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ScraperScreen(vm: ScraperViewModel = viewModel()) {
    val state by vm.state.collectAsStateWithLifecycle()
    var selectedSource by remember { mutableStateOf("all") }
    var expanded by remember { mutableStateOf(false) }

    Column(modifier = Modifier.fillMaxSize().padding(16.dp)) {
        // Controls
        Card(modifier = Modifier.fillMaxWidth()) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("Start Scrape", style = MaterialTheme.typography.titleMedium)
                Spacer(modifier = Modifier.height(12.dp))
                Row(
                    horizontalArrangement = Arrangement.spacedBy(12.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    ExposedDropdownMenuBox(
                        expanded = expanded,
                        onExpandedChange = { expanded = it },
                        modifier = Modifier.weight(1f)
                    ) {
                        OutlinedTextField(
                            value = sources.first { it.first == selectedSource }.second,
                            onValueChange = {},
                            readOnly = true,
                            trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded) },
                            modifier = Modifier.menuAnchor()
                        )
                        ExposedDropdownMenu(
                            expanded = expanded,
                            onDismissRequest = { expanded = false }
                        ) {
                            sources.forEach { (value, label) ->
                                DropdownMenuItem(
                                    text = { Text(label) },
                                    onClick = { selectedSource = value; expanded = false }
                                )
                            }
                        }
                    }
                    Button(
                        onClick = { vm.startScrape(selectedSource) },
                        enabled = !state.isRunning
                    ) {
                        if (state.isRunning) {
                            CircularProgressIndicator(modifier = Modifier.size(18.dp), strokeWidth = 2.dp)
                            Spacer(modifier = Modifier.width(8.dp))
                            Text("Running...")
                        } else {
                            Icon(Icons.Default.PlayArrow, contentDescription = null, modifier = Modifier.size(18.dp))
                            Spacer(modifier = Modifier.width(4.dp))
                            Text("Start")
                        }
                    }
                }

                if (state.isRunning) {
                    Spacer(modifier = Modifier.height(8.dp))
                    LinearProgressIndicator(modifier = Modifier.fillMaxWidth())
                }
            }
        }

        Spacer(modifier = Modifier.height(16.dp))

        // History
        Text("Scrape History", style = MaterialTheme.typography.titleMedium)
        Spacer(modifier = Modifier.height(8.dp))

        LazyColumn {
            items(state.runs) { run ->
                Card(
                    modifier = Modifier.fillMaxWidth().padding(vertical = 3.dp),
                    colors = CardDefaults.cardColors(
                        containerColor = when (run.status) {
                            "completed" -> MaterialTheme.colorScheme.secondaryContainer
                            "running" -> MaterialTheme.colorScheme.primaryContainer
                            "failed" -> MaterialTheme.colorScheme.errorContainer
                            else -> MaterialTheme.colorScheme.surfaceVariant
                        }
                    )
                ) {
                    Row(
                        modifier = Modifier.padding(12.dp).fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Column {
                            Text(
                                run.source.replace("_", " "),
                                style = MaterialTheme.typography.bodyMedium
                            )
                            Text(
                                "${run.status} | ${run.products_found} found, ${run.products_saved} saved",
                                fontSize = 11.sp,
                                color = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }
                        Column(horizontalAlignment = Alignment.End) {
                            if (run.errors_count > 0) {
                                Text("${run.errors_count} errors", fontSize = 11.sp, color = MaterialTheme.colorScheme.error)
                            }
                            val duration = if (run.started_at != null && run.completed_at != null) {
                                try {
                                    val start = java.time.Instant.parse(run.started_at)
                                    val end = java.time.Instant.parse(run.completed_at)
                                    "${java.time.Duration.between(start, end).seconds}s"
                                } catch (_: Exception) { "" }
                            } else if (run.status == "running") "..." else ""
                            if (duration.isNotEmpty()) {
                                Text(duration, fontSize = 11.sp, color = MaterialTheme.colorScheme.onSurfaceVariant)
                            }
                        }
                    }
                }
            }
        }
    }
}
