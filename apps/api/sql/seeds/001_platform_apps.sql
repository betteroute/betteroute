-- Seed: Platform Apps
-- Popular apps with deep link schemes and store metadata.
-- Run after schema migrations: psql -f sql/seeds/001_platform_apps.sql

INSERT INTO platform_apps (id, name, url_patterns, ios_scheme, android_scheme, ios_app_id, ios_bundle_id, android_package) VALUES

-- Video & Streaming
('app_youtube',       'YouTube',        '{"youtube.com","youtu.be","m.youtube.com"}',            'youtube://{path}',         'vnd.youtube://{path}',        '544007664',  'com.google.ios.youtube',       'com.google.android.youtube'),
('app_tiktok',        'TikTok',         '{"tiktok.com","vm.tiktok.com"}',                        'snssdk1233://{path}',      'snssdk1180://{path}',         '835599320',  'com.zhiliaoapp.musically',     'com.zhiliaoapp.musically'),
('app_netflix',       'Netflix',        '{"netflix.com","www.netflix.com"}',                      'nflx://{path}',            'nflx://{path}',               '363590051',  'com.netflix.Netflix',          'com.netflix.mediaclient'),

-- Music & Audio
('app_spotify',       'Spotify',        '{"open.spotify.com","spotify.com"}',                    'spotify://{path}',         'spotify://{path}',            '324684580',  'com.spotify.client',           'com.spotify.music'),
('app_apple_music',   'Apple Music',    '{"music.apple.com","itunes.apple.com"}',                 'music://{path}',           NULL,                          '1108187390', 'com.apple.music',              NULL),
('app_soundcloud',    'SoundCloud',     '{"soundcloud.com","on.soundcloud.com"}',                 'soundcloud://{path}',      'soundcloud://{path}',         '336353151',  'com.soundcloud.TouchApp',      'com.soundcloud.android'),

-- Social
('app_instagram',     'Instagram',      '{"instagram.com","www.instagram.com"}',                  'instagram://{path}',       'instagram://{path}',          '389801252',  'com.burbn.instagram',          'com.instagram.android'),
('app_twitter',       'X (Twitter)',    '{"twitter.com","x.com","t.co"}',                         'twitter://{path}',         'twitter://{path}',            '333903271',  'com.atebits.Tweetie2',         'com.twitter.android'),
('app_facebook',      'Facebook',       '{"facebook.com","fb.com","m.facebook.com"}',              'fb://{path}',              'fb://{path}',                 '284882215',  'com.facebook.Facebook',        'com.facebook.katana'),
('app_linkedin',      'LinkedIn',       '{"linkedin.com","www.linkedin.com"}',                    'linkedin://{path}',        'linkedin://{path}',           '288429040',  'com.linkedin.LinkedIn',        'com.linkedin.android'),
('app_snapchat',      'Snapchat',       '{"snapchat.com","www.snapchat.com"}',                    'snapchat://{path}',        'snapchat://{path}',           '447188370',  'com.toyopagroup.picaboo',      'com.snapchat.android'),
('app_pinterest',     'Pinterest',      '{"pinterest.com","pin.it"}',                             'pinterest://{path}',       'pinterest://{path}',          '429047995',  'com.pinterest',                'com.pinterest'),
('app_reddit',        'Reddit',         '{"reddit.com","www.reddit.com","redd.it"}',              'reddit://{path}',          'reddit://{path}',             '1064216828', 'com.reddit.Reddit',            'com.reddit.frontpage'),
('app_threads',       'Threads',        '{"threads.net","www.threads.net"}',                      'barcelona://{path}',       NULL,                          '6446901002', 'com.burbn.barcelona',          'com.instagram.barcelona'),

-- Messaging
('app_whatsapp',      'WhatsApp',       '{"wa.me","api.whatsapp.com","whatsapp.com"}',            'whatsapp://{path}',        'whatsapp://{path}',           '310633997',  'net.whatsapp.WhatsApp',        'com.whatsapp'),
('app_telegram',      'Telegram',       '{"t.me","telegram.me","telegram.org"}',                  'tg://{path}',              'tg://{path}',                 '686449807',  'ph.telegra.Telegraph',         'org.telegram.messenger'),
('app_discord',       'Discord',        '{"discord.com","discord.gg","discordapp.com"}',          'discord://{path}',         'discord://{path}',            '985746746',  'com.hammerandchisel.discord',  'com.discord'),

-- Shopping
('app_amazon',        'Amazon',         '{"amazon.com","amzn.to","amazon.co.uk","amazon.de","amazon.in"}', 'com.amazon.mobile.shopping://{path}', 'com.amazon.mobile.shopping://{path}', '297606951', 'com.amazon.Amazon', 'com.amazon.mShop.android.shopping'),
('app_flipkart',      'Flipkart',       '{"flipkart.com","dl.flipkart.com"}',                     'flpk://{path}',            'flipkart://{path}',           '742044692',  'com.flipkart.app',             'com.flipkart.android'),

-- Travel & Rides
('app_airbnb',        'Airbnb',         '{"airbnb.com","www.airbnb.com"}',                        'airbnb://{path}',          'airbnb://{path}',             '401626263',  'com.airbnb.app',               'com.airbnb.android'),
('app_uber',          'Uber',           '{"uber.com","m.uber.com"}',                              'uber://{path}',            'uber://{path}',               '368677368',  'com.ubercab.UberClient',       'com.ubercab'),

-- Productivity
('app_notion',        'Notion',         '{"notion.so","www.notion.so"}',                           'notion://{path}',          'notion://{path}',             '1232780281', 'notion.id',                    'notion.id'),
('app_figma',         'Figma',          '{"figma.com","www.figma.com"}',                           'figma://{path}',           'figma://{path}',              '1152747299', 'com.figma.FigmaPrototype',     'com.figma.mirror'),
('app_google_maps',   'Google Maps',    '{"maps.google.com","goo.gl/maps"}',                       'comgooglemaps://{path}',   'google.navigation://{path}',  '585027354',  'com.google.Maps',              'com.google.android.apps.maps'),
('app_zoom',          'Zoom',           '{"zoom.us","us02web.zoom.us"}',                           'zoomus://{path}',          'zoomus://{path}',             '546505307',  'us.zoom.videomeetings',        'us.zoom.videomeetings'),
('app_github',        'GitHub',         '{"github.com","gist.github.com"}',                        'github://{path}',          'github://{path}',             '1477376905', 'com.github.stormcrow',         'com.github.android'),

-- News & Reading
('app_medium',        'Medium',         '{"medium.com"}',                                          'medium://{path}',          'medium://{path}',             '828256236',  'com.medium.reader',            'com.medium.reader'),
('app_substack',      'Substack',       '{"substack.com"}',                                        'substack://{path}',        'substack://{path}',           '1234617060', 'com.substack.app',             'com.substack.app'),

-- Finance
('app_paypal',        'PayPal',         '{"paypal.com","paypal.me"}',                              'paypal://{path}',          'paypal://{path}',             '283646709',  'com.paypal.PPClient',          'com.paypal.android.p2pmobile')

ON CONFLICT (id) DO UPDATE SET
    name            = EXCLUDED.name,
    url_patterns    = EXCLUDED.url_patterns,
    ios_scheme      = EXCLUDED.ios_scheme,
    android_scheme  = EXCLUDED.android_scheme,
    ios_app_id      = EXCLUDED.ios_app_id,
    ios_bundle_id   = EXCLUDED.ios_bundle_id,
    android_package = EXCLUDED.android_package,
    updated_at      = NOW();
