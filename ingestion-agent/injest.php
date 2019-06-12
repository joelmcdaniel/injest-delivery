<?php 
// Get config values for redis
include 'config.php';
$host = $config['REDIS_HOST'];
$port = $config['REDIS_PORT'];
$pwd = $config['REDIS_PWD'];
$queueName = $config['QUEUE_NAME'];

// "include" for Postback class instantiation
spl_autoload_register(function ($class_name) {
    require_once $class_name . '.php';
});

// Handle web request and validate raw post data
$request;
(function(&$request){
    // Make sure that the content type of the request has been set to application/json
    $contentType = isset($_SERVER["CONTENT_TYPE"]) ? trim($_SERVER["CONTENT_TYPE"]) : '';
    if(strcasecmp($contentType, 'application/json') != 0)
        die("Content type must be 'application/json'");

    $request = json_decode(trim(file_get_contents( 'php://input' )), true);
    if(!is_array($request))
        die("Content received contained invalid json.");

    if (!isset($request['endpoint']) || !isset($request['data']))
        die("Raw post data is missing either of expected 'endpoint' or 'data' fields.");

})($request);

$endpoint = $request['endpoint'];
$data = $request['data'];

// Match potential data object property names that occur in url within {}
preg_match_all("/{(.*?)}/", $endpoint['url'], $matches);
$propNames = $matches[1];   // <-these will be replaced in url with matching property values

// Callback to create a postback object (from array_map below)
$postback = function($data) use($endpoint, $propNames) {

    // Transform url replacing {property name} with existing property value
    $url = array_reduce($propNames, function($url, $propName) use ($data) {
            $value = isset($data[$propName]) ? urlencode($data[$propName]) : "";
            return preg_replace("/{" . $propName . "}/", $value, $url);            
        }, $endpoint['url']);

    // Postback object containing http method and transformed url
    return new Postback($endpoint['method'], $url);
};

// Build an array of postback objects from each data object in data array
if (count($data) > 0) {
    $postbacks = array_map($postback, $data);

    // Connect to redis and enqueue postback objects
    if (count($postbacks) > 0) {
        $redis = new Redis();
        $connected = $redis->pconnect($host, $port);
        $auth = $redis->auth($pwd);

        if ($connected) {
            foreach ($postbacks as $postback) {
                $redis->lpush($queueName, json_encode($postback));
            }
        }
        else
            die( "Cannot connect to redis server.");
    }
}
?> 