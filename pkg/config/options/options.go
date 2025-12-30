package options

type ConfigOption string

// COBRA ONLY MISCELLANEOUS OPTIONS
const (
	// ConfigPath Path to the configuration file
	ConfigPath ConfigOption = "config-path"
	Debug      ConfigOption = "debug"
	Quiet      ConfigOption = "quiet"
)

const (
	LoggerKeyPadding ConfigOption = "key_padding"
	LoggerLevel      ConfigOption = "level"
	LoggerType       ConfigOption = "type"

	NetDomainBlacklist          ConfigOption = "domain_blacklist"
	NetDomainWhitelist          ConfigOption = "domain_whitelist"
	NetDownloadCallbackInterval ConfigOption = "download_callback_interval"
	NetExtraHeaders             ConfigOption = "extra_headers"
	NetRequestTimeout           ConfigOption = "request_timeout"
	NetUserAgent                ConfigOption = "user_agent"

	RelaySinks ConfigOption = "relay_sinks"

	ReleaserAllowAnyExtension  ConfigOption = "allow_any_extension"
	ReleaserFilePattern        ConfigOption = "file_pattern"
	ReleaserGenerateChecksums  ConfigOption = "generate_checksums"
	ReleaserGenerateSignatures ConfigOption = "generate_signatures"
	ReleaserOutputPath         ConfigOption = "output_path"
	ReleaserTargetPath         ConfigOption = "target_path"
	ReleaserPrivateKey         ConfigOption = "private_key"
	ReleaserPrivateKeyPath     ConfigOption = "private_key_path"
	ReleaserStrict             ConfigOption = "strict"
	ReleaserRequireVersion     ConfigOption = "require_version"
	ReleaserSummaryOutputType  ConfigOption = "summary_output_type"
	ReleaserVersion            ConfigOption = "version"

	UpdaterAllowDowngrade  ConfigOption = "allow_downgrade"
	UpdaterAllowPrerelease ConfigOption = "allow_prerelease"
	UpdaterArchitecture    ConfigOption = "architecture"
	UpdaterCheckInterval   ConfigOption = "check_interval"
	UpdaterCurrentVersion  ConfigOption = "current_version"
	UpdaterLogPath         ConfigOption = "log_path"
	UpdaterPlatform        ConfigOption = "platform"
	UpdaterPublicKey       ConfigOption = "public_key"
	UpdaterPublicKeyPath   ConfigOption = "public_key_path"
	UpdaterTemporaryPath   ConfigOption = "temporary_path"
	UpdaterVariant         ConfigOption = "variant"
)
