<?php

class Postback {

    public $method; // HTTP method
    public $url;    // endpoint url
  
    public function __construct($method, $url) {
        $this->method = $method;
        $this->url = $url;
    }
}

?>