#include <string>
#include <stdint.h>

class ISteamClient017;
class IClientEngine;
class ISteamUser017;

class Steam
{
public:
	static std::wstring GetSteamInstallFolder();

	ISteamUser017* GetSteamUser();

	int Initiate();

private:
	ISteamClient017* steamClient;
	ISteamUser017* steamUser;

	int32_t steamPipeHandle;
	int32_t steamUserHandle;

};
