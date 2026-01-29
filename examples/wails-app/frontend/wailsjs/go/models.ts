export namespace main {
	
	export enum Channel {
	    RELAY_NET = "net",
	    RELAY_BASE = "relay",
	    RELAY_RELEASER = "releaser",
	    RELAY_UPDATER = "updater",
	}
	export enum Relay {
	    NET_DOWNLOAD = "net.download",
	    NET_LOG = "net.log",
	    RELAY_LOG = "relay.log",
	    RELEASE_LOG = "release.log",
	    UPDATER_LOG = "updater.log",
	    UPDATER_NEW_VERSION = "updater.new_version",
	}

}

export namespace releaserdto {
	
	export interface ReleaseAsset {
	    artefact_name: string;
	    platform: string;
	    arch: string;
	    variant: string;
	    version: string;
	    download_url: string;
	    checksum: string;
	    size_bytes: number;
	    signature?: string;
	    signature_type?: string;
	}

}

export namespace updaterdto {
	
	export interface UpdaterState {
	    updater_last_time_checked_update?: string;
	    updater_update_link?: releaserdto.ReleaseAsset;
	    updater_changelog: string;
	    updater_released_at?: string;
	    updater_check_interval: number;
	    updater_log: string;
	    updater_log_path: string;
	    updater_updating: boolean;
	    updater_version: string;
	}

}

