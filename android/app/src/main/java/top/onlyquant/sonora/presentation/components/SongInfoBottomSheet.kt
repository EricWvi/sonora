@file:OptIn(ExperimentalMaterial3ExpressiveApi::class)

package top.onlyquant.sonora.presentation.components

import android.app.Activity
import android.content.ActivityNotFoundException
import android.content.Context
import android.content.Intent
import android.net.Uri
import android.provider.Settings
import android.widget.Toast
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.animateContentSize
import androidx.compose.animation.core.animateDpAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.LocalOverscrollFactory
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.pager.HorizontalPager
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.QueueMusic
import androidx.compose.material.icons.automirrored.rounded.PlaylistAdd
import androidx.compose.material.icons.automirrored.rounded.QueueMusic
import androidx.compose.material.icons.filled.DeleteForever
import androidx.compose.material.icons.filled.Favorite
import androidx.compose.material.icons.filled.FavoriteBorder
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.QueueMusic
import androidx.compose.material.icons.rounded.Album
import androidx.compose.material.icons.rounded.AudioFile
import androidx.compose.material.icons.rounded.Cloud
import androidx.compose.material.icons.rounded.CloudDownload
import androidx.compose.material.icons.rounded.CloudDone
import androidx.compose.material.icons.rounded.Favorite
import androidx.compose.material.icons.rounded.FavoriteBorder
import androidx.compose.material.icons.rounded.MusicNote
import androidx.compose.material.icons.rounded.Person
import androidx.compose.material.icons.rounded.PlayArrow
import androidx.compose.material.icons.rounded.Schedule
import androidx.compose.material.icons.rounded.Share
import androidx.compose.material.icons.rounded.Edit
import androidx.compose.material.icons.rounded.Info
import androidx.compose.material.icons.rounded.Menu
import androidx.compose.material.icons.rounded.TravelExplore
import androidx.compose.material3.ExperimentalMaterial3ExpressiveApi
import androidx.compose.material3.LoadingIndicator
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Shape
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.core.content.FileProvider
import top.onlyquant.sonora.R
import top.onlyquant.sonora.data.model.Song
import top.onlyquant.sonora.presentation.components.subcomps.AutoSizingTextToFill
import top.onlyquant.sonora.utils.formatDuration
import racra.compose.smooth_corner_rect_library.AbsoluteSmoothCornerShape
import androidx.core.net.toUri
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import kotlinx.coroutines.launch

import androidx.compose.ui.graphics.TransformOrigin
import top.onlyquant.sonora.presentation.screens.TabAnimation
import top.onlyquant.sonora.ui.theme.RoundedSans
import top.onlyquant.sonora.utils.AudioMetaUtils
import androidx.compose.ui.res.stringResource


/**
 * From SongInfoBottomSheet in PixelPlayerOSS v0.4.0-alpha.53 (GPL-3.0), the song's "⋮" menu.
 * Sonora (Eric, 2026-11-11): only the options page, with play, like, add to queue, play next and add
 * to playlist. The original also has share, delete (or download, for cloud songs), set as
 * ringtone, MusicBrainz lookup, an edit button next to the title and a details page; without the
 * details page there is no options/details switch at the bottom.
 */
@OptIn(ExperimentalMaterial3Api::class, ExperimentalMaterial3ExpressiveApi::class)
@Composable
fun SongInfoBottomSheet(
    song: Song,
    isFavorite: Boolean,
    onToggleFavorite: () -> Unit,
    onDismiss: () -> Unit,
    onPlaySong: () -> Unit,
    onAddToQueue: () -> Unit,
    onAddNextToQueue: () -> Unit,
    onAddToPlayList: () -> Unit
) {
    val evenCornerRadiusElems = 26.dp

    val listItemShape = remember {
        AbsoluteSmoothCornerShape(
            cornerRadiusTR = 20.dp, smoothnessAsPercentBR = 60, cornerRadiusBR = 20.dp,
            smoothnessAsPercentTL = 60, cornerRadiusTL = 20.dp, smoothnessAsPercentBL = 60,
            cornerRadiusBL = 20.dp, smoothnessAsPercentTR = 60
        )
    }
    val albumArtShape = remember(evenCornerRadiusElems) {
        AbsoluteSmoothCornerShape(
            cornerRadiusTR = evenCornerRadiusElems, smoothnessAsPercentBR = 60, cornerRadiusBR = evenCornerRadiusElems,
            smoothnessAsPercentTL = 60, cornerRadiusTL = evenCornerRadiusElems, smoothnessAsPercentBL = 60,
            cornerRadiusBL = evenCornerRadiusElems, smoothnessAsPercentTR = 60
        )
    }
    val playButtonShape = remember(evenCornerRadiusElems) {
        AbsoluteSmoothCornerShape(
            cornerRadiusTR = evenCornerRadiusElems, smoothnessAsPercentBR = 60, cornerRadiusBR = evenCornerRadiusElems,
            smoothnessAsPercentTL = 60, cornerRadiusTL = evenCornerRadiusElems, smoothnessAsPercentBL = 60,
            cornerRadiusBL = evenCornerRadiusElems, smoothnessAsPercentTR = 60
        )
    }

    val sheetState = rememberModalSheetState(
        skipPartiallyExpanded = true,
        confirmValueChange = { true }
    )

    val favoriteButtonCornerRadius by animateDpAsState(
        targetValue = if (isFavorite) evenCornerRadiusElems else 60.dp,
        animationSpec = tween(durationMillis = 300), label = "FavoriteCornerAnimation"
    )
    val favoriteButtonContainerColor by animateColorAsState(
        targetValue = if (isFavorite) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.surfaceVariant,
        animationSpec = tween(durationMillis = 300), label = "FavoriteContainerColorAnimation"
    )
    val favoriteButtonContentColor by animateColorAsState(
        targetValue = if (isFavorite) MaterialTheme.colorScheme.onPrimary else MaterialTheme.colorScheme.onSurfaceVariant,
        animationSpec = tween(durationMillis = 300), label = "FavoriteContentColorAnimation"
    )
    val favoriteButtonShape = remember(favoriteButtonCornerRadius) {
        AbsoluteSmoothCornerShape(
            cornerRadiusTR = favoriteButtonCornerRadius, smoothnessAsPercentBR = 60, cornerRadiusBR = favoriteButtonCornerRadius,
            smoothnessAsPercentTL = 60, cornerRadiusTL = favoriteButtonCornerRadius, smoothnessAsPercentBL = 60,
            cornerRadiusBL = favoriteButtonCornerRadius, smoothnessAsPercentTR = 60
        )
    }

    val configuration = androidx.compose.ui.platform.LocalConfiguration.current
    val safeInsets = WindowInsets.safeDrawing.asPaddingValues()
    // Keep the pager from growing past the screen on small devices: the sheet header and
    // grab handle take ~180dp; never shrink below 280dp so content stays usable.
    val maxPagerHeight = (
        configuration.screenHeightDp.dp -
            safeInsets.calculateTopPadding() -
            safeInsets.calculateBottomPadding() -
            180.dp
        ).coerceAtLeast(280.dp)

    ModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = sheetState,
    ) {
        CompositionLocalProvider(
            LocalOverscrollFactory provides null
        ) {
            Box(
                modifier = Modifier
                    .fillMaxWidth(),
                contentAlignment = Alignment.TopCenter
            ) {
                Column(
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Column(
                        modifier = Modifier
                            .padding(horizontal = 16.dp)
                    ) {
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(80.dp),
                            horizontalArrangement = Arrangement.spacedBy(14.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            SmartImage(
                                model = song.albumArtUriString,
                                contentDescription = stringResource(R.string.widget_album_art),
                                shape = albumArtShape,
                                modifier = Modifier.size(80.dp),
                                contentScale = ContentScale.Crop
                            )
                            Box(
                                modifier = Modifier
                                    .weight(1f)
                                    .fillMaxHeight(),
                                contentAlignment = Alignment.CenterStart
                            ) {
                                AutoSizingTextToFill(
                                    modifier = Modifier.padding(end = 4.dp),
                                    fontWeight = FontWeight.Light,
                                    text = song.title
                                )
                            }
                            // Sonora: the original has an edit button here.
                        }
                    }

                    Spacer(modifier = Modifier.height(16.dp))

                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .heightIn(max = maxPagerHeight)
                    ) {
                                    LazyColumn(
                                        modifier = Modifier
                                            .fillMaxWidth(),
                                        contentPadding = PaddingValues(horizontal = 16.dp),
                                        verticalArrangement = Arrangement.spacedBy(10.dp)
                                    ) {
                                        item {
                                            Row(
                                                modifier = Modifier
                                                    .fillMaxWidth()
                                                    .height(IntrinsicSize.Min),
                                                verticalAlignment = Alignment.CenterVertically,
                                                horizontalArrangement = Arrangement.spacedBy(10.dp)
                                            ) {
                                                // Sonora: without the share button, play and like split the row
                                                // evenly and like is a labelled pill like play (Eric, 2026-11-11);
                                                // the original has play 0.5, an icon-only like 0.25 and share 0.25.
                                                MediumExtendedFloatingActionButton(
                                                    modifier = Modifier
                                                        .weight(0.5f)
                                                        .fillMaxHeight(),
                                                    onClick = onPlaySong,
                                                    elevation = FloatingActionButtonDefaults.elevation(0.dp),
                                                    shape = playButtonShape,
                                                    icon = {
                                                        Icon(Icons.Rounded.PlayArrow, contentDescription = stringResource(R.string.cd_play_song_action))
                                                    },
                                                    text = {
                                                        Text(
                                                            modifier = Modifier.padding(end = 10.dp),
                                                            text = stringResource(R.string.play_playback)
                                                        )
                                                    }
                                                )

                                                MediumExtendedFloatingActionButton(
                                                    modifier = Modifier
                                                        .weight(0.5f)
                                                        .fillMaxHeight(),
                                                    onClick = onToggleFavorite,
                                                    elevation = FloatingActionButtonDefaults.elevation(0.dp),
                                                    shape = favoriteButtonShape,
                                                    containerColor = favoriteButtonContainerColor,
                                                    contentColor = favoriteButtonContentColor,
                                                    icon = {
                                                        Icon(
                                                            imageVector = if (isFavorite) Icons.Rounded.Favorite else Icons.Rounded.FavoriteBorder,
                                                            contentDescription = stringResource(
                                                                if (isFavorite) R.string.cd_remove_from_favorites else R.string.cd_add_to_favorites
                                                            )
                                                        )
                                                    },
                                                    text = {
                                                        Text(
                                                            modifier = Modifier.padding(end = 10.dp),
                                                            text = stringResource(R.string.action_like)
                                                        )
                                                    }
                                                )
                                                // Sonora: the original has a share button here.
                                            }
                                        }
                                        item {
                                            Row(
                                                modifier = Modifier
                                                    .fillMaxWidth()
                                                    .height(IntrinsicSize.Min),
                                                verticalAlignment = Alignment.CenterVertically,
                                                horizontalArrangement = Arrangement.spacedBy(10.dp)
                                            ) {
                                                FilledTonalButton(
                                                    modifier = Modifier
                                                        .weight(0.6f)
                                                        .heightIn(min = 66.dp),
                                                    colors = ButtonDefaults.filledTonalButtonColors(
                                                        containerColor = MaterialTheme.colorScheme.tertiaryContainer,
                                                        contentColor = MaterialTheme.colorScheme.onTertiaryContainer
                                                    ),
                                                    contentPadding = PaddingValues(horizontal = 0.dp),
                                                    shape = CircleShape,
                                                    onClick = onAddToQueue
                                                ) {
                                                    Icon(
                                                        Icons.AutoMirrored.Rounded.QueueMusic,
                                                        contentDescription = stringResource(R.string.cd_add_to_queue)
                                                    )
                                                    Spacer(Modifier.width(14.dp))
                                                    Text(stringResource(R.string.action_add_to_queue))
                                                }
                                                FilledTonalButton(
                                                    modifier = Modifier
                                                        .weight(0.4f)
                                                        .heightIn(min = 66.dp),
                                                    colors = ButtonDefaults.filledTonalButtonColors(
                                                        containerColor = MaterialTheme.colorScheme.tertiary,
                                                        contentColor = MaterialTheme.colorScheme.onTertiary
                                                    ),
                                                    contentPadding = PaddingValues(horizontal = 0.dp),
                                                    shape = CircleShape,
                                                    onClick = onAddNextToQueue
                                                ) {
                                                    Icon(
                                                        Icons.AutoMirrored.Filled.QueueMusic,
                                                        contentDescription = stringResource(R.string.cd_play_next_in_queue)
                                                    )
                                                    Spacer(Modifier.width(8.dp))
                                                    Text(stringResource(R.string.action_queue_next))
                                                }
                                            }
                                        }

                                        item {
                                            Row(
                                                modifier = Modifier
                                                    .fillMaxWidth()
                                                    .height(IntrinsicSize.Min),
                                                verticalAlignment = Alignment.CenterVertically,
                                                horizontalArrangement = Arrangement.spacedBy(10.dp)
                                            ) {
                                                FilledTonalButton(
                                                    modifier = Modifier
                                                        .weight(0.5f)
                                                        .heightIn(min = 66.dp),
                                                    colors = ButtonDefaults.filledTonalButtonColors(
                                                        containerColor = MaterialTheme.colorScheme.secondaryContainer,
                                                        contentColor = MaterialTheme.colorScheme.onSecondaryContainer
                                                    ),
                                                    shape = CircleShape,
                                                    onClick = onAddToPlayList
                                                ) {
                                                    Icon(
                                                        Icons.AutoMirrored.Rounded.PlaylistAdd,
                                                        contentDescription = stringResource(R.string.cd_add_to_playlist)
                                                    )
                                                    Spacer(Modifier.width(8.dp))
                                                    Text(stringResource(R.string.shortcut_playlist_short))
                                                }
                                                // Sonora: the original has delete (download for
                                                // cloud songs) here, then set as ringtone and the
                                                // MusicBrainz lookup below.
                                            }
                                        }

                                        // The original keeps this space for its options/details
                                        // switch; Sonora keeps it so the sheet has the same height.
                                        item {
                                            Spacer(Modifier.height(80.dp))
                                        }
                                    }
                    }
                }
            }
        }
    }
}
