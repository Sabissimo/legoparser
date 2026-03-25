package com.legoparser.ui.navigation

import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.CompareArrows
import androidx.compose.material.icons.filled.Inventory2
import androidx.compose.material.icons.filled.Sync
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import com.legoparser.ui.screens.ComparisonScreen
import com.legoparser.ui.screens.ProductsScreen
import com.legoparser.ui.screens.ScraperScreen

enum class Screen(val route: String, val label: String, val icon: ImageVector) {
    Products("products", "Products", Icons.Default.Inventory2),
    Comparison("comparison", "Prices", Icons.Default.CompareArrows),
    Scraper("scraper", "Scraper", Icons.Default.Sync),
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AppNavigation() {
    val nav = rememberNavController()
    val current = nav.currentBackStackEntryAsState().value?.destination?.route

    Scaffold(
        topBar = { TopAppBar(title = { Text("LEGO Parser") }, colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.primaryContainer)) },
        bottomBar = {
            NavigationBar {
                Screen.entries.forEach { s ->
                    NavigationBarItem(
                        selected = current == s.route,
                        onClick = { nav.navigate(s.route) { popUpTo(nav.graph.startDestinationId) { saveState = true }; launchSingleTop = true; restoreState = true } },
                        icon = { Icon(s.icon, s.label) },
                        label = { Text(s.label) }
                    )
                }
            }
        }
    ) { padding ->
        NavHost(nav, Screen.Products.route, Modifier.padding(padding)) {
            composable(Screen.Products.route) { ProductsScreen() }
            composable(Screen.Comparison.route) { ComparisonScreen() }
            composable(Screen.Scraper.route) { ScraperScreen() }
        }
    }
}
