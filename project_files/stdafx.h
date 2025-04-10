#pragma once
#define COOPANDREAS_VERSION "0.1.1-alpha"
#define MAX_SERVER_PLAYERS 4
#define MAX_SERVER_VEHICLES 200
#define MAX_SERVER_PEDS 300
#define MIN_SCREEN_WIDTH 640
#define MIN_SCREEN_HEIGHT 480
#define MAX_SCREEN_WIDTH 2560
#define MAX_SCREEN_HEIGHT 1600
#define MIN_RENDER_FONT_SIZE 12
#define MAX_RENDER_FONT_SIZE 23
#define _DEV
//#define PACKET_DEBUG_MESSAGES
#define PURECALL (0x82263A)

#define ARRAY_SIZE(a) (sizeof(a) / sizeof(a[0]))

// networking ------------------------------------------------------------------------------

#include "enet/enet.h"

// windows ---------------------------------------------------------------------------------

#include <Windows.h>

// utf8 ------------------------------------------------------------------------------------

#include <cwctype>
#include "utf8.h"

// plugin-sdk ------------------------------------------------------------------------------

#include "plugin.h"
#include "Patch.h"
#include "CPlayerPed.h"
#include "CStreaming.h"
#include "CWorld.h"
#include "RenderWare.h"
#include "d3dx9.h"
#include "extensions/ScriptCommands.h"
#include "CTaskSimpleDuckToggle.h"
#include "CTaskSimplePlayerOnFoot.h"
#include "CRadar.h"
#include "CText.h"
#include "CCamera.h"
#include "CBulletTraces.h"
#include "CDraw.h"
#include "CMenuManager.h"
#include "CSprite.h"
#include "CFont.h"
#include "CHud.h"
#include "CModelInfo.h"
#include "CTxdStore.h"
#include "CKeyGen.h"
#include "CAnimManager.h"
#include "CMessages.h"
#include "CExplosion.h"
#include "CMonsterTruck.h"
#include "CQuadBike.h"
#include "CHeli.h"
#include "CPlane.h"
#include "CBmx.h"
#include "CTrailer.h"
#include "CTrain.h"
#include "CBoat.h"
#include "CTaskComplexEnterCarAsDriver.h"
#include "CTaskComplexLeaveCar.h"
#include "CCarCtrl.h"
#include "CCheat.h"
#include "CEmergencyPed.h"
#include "CCivilianPed.h"
#include "CTaskComplexWanderStandard.h"
#include "CTheScripts.h"
#include "CTaskSimpleGangDriveBy.h"
#include "CLoadingScreen.h"
#include "CPopulation.h"
#include "CWeather.h"
#include "CClock.h"
#include <CTaskComplexWander.h>
#include <CTaskComplexClimb.h>
#include <CTaskSimpleCarSetPedInAsDriver.h>
#include <CStats.h>
#include <CClothes.h>
#include <CCutsceneMgr.h>
#include <CGame.h>
#include <CAudioEngine.h>
#include "game_sa/CTaskComplexCarDriveWander.h"

// classes ---------------------------------------------------------------------------------

using namespace plugin;

#include "CCore.h"
#include "CUnicode.h"
#include "CDXFont.h"
#include "structs.h"
#include "CNetworkVehicle.h"
#include "CNetworkPed.h"
#include "CPackets.h"
#include "CPacketHandler.h"
#include "CNetwork.h"
#include "CNetworkPlayerStats.h"
#include "CNetworkPlayer.h"
#include "CNetworkPlayerManager.h"
#include "CUtil.h"
#include "CPatch.h"
#include "Hooks/CHook.h"
#include "Events.h"
#include "CLocalPlayer.h"
#include "CNetworkPlayerMapPin.h"
#include "CNetworkPlayerWaypoint.h"
#include "CLaunchManager.h"
#include "CNetworkPlayerNameTag.h"
#include "CChat.h"
#include "CNetworkVehicleManager.h"
#include "CDebugVehicleSpawner.h"
#include "CPassengerEnter.h"
#include "CDriveBy.h"
#include "CNetworkPedManager.h"
#include "CDebugPedTasks.h"
#include "eNetworkEntityType.h"
#include "CStatsSync.h"
#include "CMassPacketBuilder.h"
// -----------------------------------------------------------------------------------------
